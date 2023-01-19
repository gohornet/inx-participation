package participation

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/core/generics/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/inx-app/pkg/httpserver"
	"github.com/iotaledger/inx-participation/pkg/participation"
	iotago "github.com/iotaledger/iota.go/v3"
)

// EventIDFromHex creates a EventID from a hex string representation.
func EventIDFromHex(hexString string) (participation.EventID, error) {

	b, err := iotago.DecodeHex(hexString)
	if err != nil {
		return participation.NullEventID, err
	}

	if len(b) != participation.EventIDLength {
		return participation.NullEventID, fmt.Errorf("unknown eventID length (%d)", len(b))
	}

	var eventID participation.EventID
	copy(eventID[:], b)

	return eventID, nil
}

func parseEventTypeQueryParam(c echo.Context) ([]uint32, error) {
	typeParams := c.QueryParams()["type"]

	if len(typeParams) == 0 {
		return []uint32{}, nil
	}

	returnTypes := make([]uint32, len(typeParams))
	for i, typeParam := range typeParams {
		intParam, err := strconv.ParseUint(typeParam, 10, 32)
		if err != nil {
			return []uint32{}, errors.WithMessagef(httpserver.ErrInvalidParameter, "invalid event type: %s, error: %s", typeParam, err)
		}

		eventType := uint32(intParam)
		switch eventType {
		case participation.BallotPayloadTypeID:
		case participation.StakingPayloadTypeID:
		default:
			return []uint32{}, errors.WithMessagef(httpserver.ErrInvalidParameter, "invalid event type: %s", typeParam)
		}

		returnTypes[i] = eventType
	}

	return returnTypes, nil
}

func parseEventIDParam(c echo.Context) (participation.EventID, error) {
	eventIDHex := strings.ToLower(c.Param(ParameterParticipationEventID))
	if eventIDHex == "" {
		return participation.NullEventID, errors.WithMessagef(httpserver.ErrInvalidParameter, "parameter \"%s\" not specified", ParameterParticipationEventID)
	}

	eventID, err := EventIDFromHex(eventIDHex)
	if err != nil {
		return participation.NullEventID, errors.WithMessagef(httpserver.ErrInvalidParameter, "invalid event ID: %s, error: %s", eventIDHex, err)
	}

	return eventID, nil
}

func getEvents(c echo.Context) (*EventsResponse, error) {

	eventTypes, err := parseEventTypeQueryParam(c)
	if err != nil {
		return nil, err
	}

	eventIDs := deps.ParticipationManager.EventIDs(eventTypes...)

	hexEventIDs := []string{}
	for _, id := range eventIDs {
		hexEventIDs = append(hexEventIDs, id.ToHex())
	}
	sort.Strings(hexEventIDs)

	return &EventsResponse{EventIDs: hexEventIDs}, nil
}

func createEvent(c echo.Context) (*CreateEventResponse, error) {

	event := &participation.Event{}
	if err := c.Bind(event); err != nil {
		return nil, errors.WithMessagef(httpserver.ErrInvalidParameter, "invalid request, error: %s", err)
	}

	if _, err := event.Serialize(serializer.DeSeriModePerformValidation, nil); err != nil {
		return nil, errors.WithMessagef(httpserver.ErrInvalidParameter, "invalid event payload, error: %s", err)
	}

	eventID, err := deps.ParticipationManager.StoreEvent(event)
	if err != nil {
		return nil, errors.WithMessagef(httpserver.ErrInvalidParameter, "invalid event, error: %s", err)
	}

	return &CreateEventResponse{
		EventID: eventID.ToHex(),
	}, nil
}

func getEvent(c echo.Context) (*participation.Event, error) {

	eventID, err := parseEventIDParam(c)
	if err != nil {
		return nil, err
	}

	event := deps.ParticipationManager.Event(eventID)
	if event == nil {
		return nil, errors.WithMessagef(echo.ErrNotFound, "event not found: %s", eventID.ToHex())
	}

	return event, nil
}

func deleteEvent(c echo.Context) error {

	eventID, err := parseEventIDParam(c)
	if err != nil {
		return err
	}

	if err = deps.ParticipationManager.DeleteEvent(eventID); err != nil {
		if errors.Is(err, participation.ErrEventNotFound) {
			return errors.WithMessagef(echo.ErrNotFound, "event not found: %s", eventID.ToHex())
		}

		return errors.WithMessagef(echo.ErrInternalServerError, "deleting event failed: %s", err)
	}

	return nil
}

func parseMilestoneIndexQueryParam(c echo.Context) (iotago.MilestoneIndex, error) {
	milestoneIndexParam := c.QueryParam(ParameterMilestoneIndex)
	if len(milestoneIndexParam) == 0 {
		return 0, nil
	}

	intParam, err := strconv.ParseUint(milestoneIndexParam, 10, 32)
	if err != nil {
		return 0, errors.WithMessagef(httpserver.ErrInvalidParameter, "invalid milestone index: %s, error: %s", milestoneIndexParam, err)
	}

	return iotago.MilestoneIndex(intParam), nil
}

func getEventStatus(c echo.Context) (*participation.EventStatus, error) {
	eventID, err := parseEventIDParam(c)
	if err != nil {
		return nil, err
	}

	milestoneIndex, err := parseMilestoneIndexQueryParam(c)
	if err != nil {
		return nil, err
	}

	var milestoneIndexFilter []iotago.MilestoneIndex
	if milestoneIndex > 0 {
		milestoneIndexFilter = append(milestoneIndexFilter, milestoneIndex)
	}

	status, err := deps.ParticipationManager.EventStatus(eventID, milestoneIndexFilter...)
	if err != nil {
		if errors.Is(err, participation.ErrEventNotFound) {
			return nil, errors.WithMessagef(echo.ErrNotFound, "event not found: %s", eventID.ToHex())
		}

		return nil, errors.WithMessagef(echo.ErrInternalServerError, "get event status failed: %s", err)
	}

	return status, nil
}

func getOutputStatus(c echo.Context) (*OutputStatusResponse, error) {
	outputID, err := httpserver.ParseOutputIDParam(c, ParameterOutputID)
	if err != nil {
		return nil, err
	}

	trackedParticipations, err := deps.ParticipationManager.ParticipationsForOutputID(outputID)
	if err != nil {
		return nil, errors.WithMessagef(echo.ErrInternalServerError, "error fetching participations: %s", err)
	}

	if len(trackedParticipations) == 0 {
		return nil, errors.WithMessagef(echo.ErrNotFound, "output not found: %s", outputID.ToHex())
	}

	response := &OutputStatusResponse{
		Participations: make(map[string]*TrackedParticipation),
	}

	for _, trackedParticipation := range trackedParticipations {
		answers, err := deps.ParticipationManager.AnswersForTrackedParticipation(trackedParticipation)
		if err != nil {
			return nil, errors.WithMessage(echo.ErrInternalServerError, err.Error())
		}

		t := &TrackedParticipation{
			BlockID:             trackedParticipation.BlockID.ToHex(),
			Amount:              trackedParticipation.Amount,
			StartMilestoneIndex: trackedParticipation.StartIndex,
			EndMilestoneIndex:   trackedParticipation.EndIndex,
			Answers:             lo.Map(answers, func(answer byte) int { return int(answer) }),
		}
		response.Participations[trackedParticipation.EventID.ToHex()] = t
	}

	return response, nil
}

func ParseBech32AddressParam(c echo.Context, prefix iotago.NetworkPrefix) (iotago.Address, error) {
	addressParam := strings.ToLower(c.Param(ParameterAddress))

	hrp, bech32Address, err := iotago.ParseBech32(addressParam)
	if err != nil {
		return nil, errors.WithMessagef(httpserver.ErrInvalidParameter, "invalid address: %s, error: %s", addressParam, err)
	}

	if hrp != prefix {
		return nil, errors.WithMessagef(httpserver.ErrInvalidParameter, "invalid bech32 address, expected prefix: %s", prefix)
	}

	return bech32Address, nil
}

func getRewardsByAddress(c echo.Context) (*participation.AddressRewards, error) {
	bech32Address, err := ParseBech32AddressParam(c, deps.NodeBridge.ProtocolParameters().Bech32HRP)
	if err != nil {
		return nil, err
	}

	milestoneIndex, err := parseMilestoneIndexQueryParam(c)
	if err != nil {
		return nil, err
	}

	if milestoneIndex > 0 {
		return deps.ParticipationManager.AddressRewards(bech32Address, milestoneIndex)
	}

	return deps.ParticipationManager.AddressRewards(bech32Address)
}

func getRewards(c echo.Context) (*participation.EventRewards, error) {
	eventID, err := parseEventIDParam(c)
	if err != nil {
		return nil, err
	}

	milestoneIndex, err := parseMilestoneIndexQueryParam(c)
	if err != nil {
		return nil, err
	}

	return deps.ParticipationManager.EventRewards(eventID, milestoneIndex)
}

func getOutputsByAddress(c echo.Context) (*AddressOutputsResponse, error) {
	bech32Address, err := ParseBech32AddressParam(c, deps.NodeBridge.ProtocolParameters().Bech32HRP)
	if err != nil {
		return nil, err
	}

	eventIDs := deps.ParticipationManager.EventIDs()

	response := &AddressOutputsResponse{
		Outputs: make(map[string]*OutputStatusResponse),
	}

	for _, eventID := range eventIDs {

		event := deps.ParticipationManager.Event(eventID)
		if event == nil {
			return nil, errors.WithMessage(echo.ErrInternalServerError, "event not found")
		}

		participations, err := deps.ParticipationManager.ParticipationsForAddress(eventID, bech32Address)
		if err != nil {
			return nil, errors.WithMessagef(echo.ErrInternalServerError, "error fetching outputs: %s", err)
		}
		for _, trackedParticipation := range participations {
			answers, err := deps.ParticipationManager.AnswersForTrackedParticipation(trackedParticipation)
			if err != nil {
				return nil, errors.WithMessage(echo.ErrInternalServerError, err.Error())
			}

			t := &TrackedParticipation{
				BlockID:             trackedParticipation.BlockID.ToHex(),
				Amount:              trackedParticipation.Amount,
				StartMilestoneIndex: trackedParticipation.StartIndex,
				EndMilestoneIndex:   trackedParticipation.EndIndex,
				Answers:             lo.Map(answers, func(answer byte) int { return int(answer) }),
			}
			outputResponse := response.Outputs[trackedParticipation.OutputID.ToHex()]
			if outputResponse == nil {
				outputResponse = &OutputStatusResponse{
					Participations: make(map[string]*TrackedParticipation),
				}
				response.Outputs[trackedParticipation.OutputID.ToHex()] = outputResponse
			}
			outputResponse.Participations[trackedParticipation.EventID.ToHex()] = t
		}
	}

	return response, nil
}

//nolint:dupl
func getActiveParticipations(c echo.Context) (*ParticipationsResponse, error) {
	eventID, err := parseEventIDParam(c)
	if err != nil {
		return nil, err
	}

	response := &ParticipationsResponse{
		Participations: make(map[string]*TrackedParticipation),
	}

	var innerErr error
	if err := deps.ParticipationManager.ForEachActiveParticipation(eventID, func(trackedParticipation *participation.TrackedParticipation) bool {
		answers, err := deps.ParticipationManager.AnswersForTrackedParticipation(trackedParticipation)
		if err != nil {
			innerErr = errors.WithMessage(echo.ErrInternalServerError, err.Error())
			return false
		}

		t := &TrackedParticipation{
			BlockID:             trackedParticipation.BlockID.ToHex(),
			Amount:              trackedParticipation.Amount,
			StartMilestoneIndex: trackedParticipation.StartIndex,
			EndMilestoneIndex:   trackedParticipation.EndIndex,
			Answers:             lo.Map(answers, func(answer byte) int { return int(answer) }),
		}
		response.Participations[trackedParticipation.OutputID.ToHex()] = t

		return true
	}); err != nil {
		return nil, errors.WithMessagef(echo.ErrInternalServerError, "error fetching active participations: %s", err)
	}
	if innerErr != nil {
		return nil, innerErr
	}

	return response, nil
}

//nolint:dupl
func getPastParticipations(c echo.Context) (*ParticipationsResponse, error) {
	eventID, err := parseEventIDParam(c)
	if err != nil {
		return nil, err
	}

	response := &ParticipationsResponse{
		Participations: make(map[string]*TrackedParticipation),
	}

	var innerErr error
	if err := deps.ParticipationManager.ForEachPastParticipation(eventID, func(trackedParticipation *participation.TrackedParticipation) bool {
		answers, err := deps.ParticipationManager.AnswersForTrackedParticipation(trackedParticipation)
		if err != nil {
			innerErr = errors.WithMessage(echo.ErrInternalServerError, err.Error())
			return false
		}

		t := &TrackedParticipation{
			BlockID:             trackedParticipation.BlockID.ToHex(),
			Amount:              trackedParticipation.Amount,
			StartMilestoneIndex: trackedParticipation.StartIndex,
			EndMilestoneIndex:   trackedParticipation.EndIndex,
			Answers:             lo.Map(answers, func(answer byte) int { return int(answer) }),
		}
		response.Participations[trackedParticipation.OutputID.ToHex()] = t

		return true
	}); err != nil {
		return nil, errors.WithMessagef(echo.ErrInternalServerError, "error fetching past participations: %s", err)
	}
	if innerErr != nil {
		return nil, innerErr
	}

	return response, nil
}
