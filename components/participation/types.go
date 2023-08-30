package participation

import (
	iotago "github.com/iotaledger/iota.go/v3"
)

// EventsResponse defines the response of a GET RouteParticipationEvents REST API call.
type EventsResponse struct {
	// The hex encoded IDs of the found events.
	EventIDs []string `json:"eventIds"`
}

// CreateEventResponse defines the response of a POST RouteParticipationEvents REST API call.
type CreateEventResponse struct {
	// The hex encoded ID of the created participation event.
	EventID string `json:"eventId"`
}

// TrackedParticipation holds the information for each tracked participation.
type TrackedParticipation struct {
	// BlockID is the ID of the block that included the transaction that created the output the participation was made.
	BlockID string `json:"blockId"`
	// Amount is the amount of tokens that were included in the output the participation was made.
	Amount uint64 `json:"amount"`
	// StartMilestoneIndex is the milestone index the participation started.
	StartMilestoneIndex iotago.MilestoneIndex `json:"startMilestoneIndex"`
	// EndMilestoneIndex is the milestone index the participation ended. 0 if the participation is still active.
	EndMilestoneIndex iotago.MilestoneIndex `json:"endMilestoneIndex"`
	// Answers holds the IDs of the answers to the questions of a ballot.
	Answers []int `json:"answers,omitempty"`
}

// OutputStatusResponse defines the response of a GET RouteOutputStatus REST API call.
type OutputStatusResponse struct {
	// Participations holds the participations that were created in the output.
	Participations map[string]*TrackedParticipation `json:"participations"`
}

// AddressOutputsResponse defines the response of a GET RouteAddressBech32Outputs REST API call.
type AddressOutputsResponse struct {
	// Outputs is a map of output status per outputID.
	Outputs map[string]*OutputStatusResponse `json:"outputs"`
}

// ParticipationsResponse defines the response of a GET RouteAdminActiveParticipations or RouteAdminPastParticipations REST API call.
type ParticipationsResponse struct {
	// Participations holds the participations that are/were tracked.
	Participations map[string]*TrackedParticipation `json:"participations"`
}
