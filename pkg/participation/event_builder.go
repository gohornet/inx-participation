package participation

import (
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
)

// NewEventBuilder creates a new EventBuilder.
func NewEventBuilder(name string, milestoneCommence iotago.MilestoneIndex, milestoneBeginHolding iotago.MilestoneIndex, milestoneEnd iotago.MilestoneIndex, additionalInfo string) *EventBuilder {
	return &EventBuilder{
		event: &Event{
			Name:                   name,
			MilestoneIndexCommence: milestoneCommence,
			MilestoneIndexStart:    milestoneBeginHolding,
			MilestoneIndexEnd:      milestoneEnd,
			AdditionalInfo:         additionalInfo,
		},
	}
}

// EventBuilder is used to easily build up a Event.
type EventBuilder struct {
	event *Event
	err   error
}

// Payload sets the payload to embed within the block.
func (rb *EventBuilder) Payload(seri serializer.Serializable) *EventBuilder {
	if rb.err != nil {
		return rb
	}
	switch seri.(type) {
	case *Ballot:
	case *Staking:
	case nil:
	default:
		rb.err = fmt.Errorf("%w: unsupported type %T", ErrUnknownPayloadType, seri)

		return rb
	}
	rb.event.Payload = seri

	return rb
}

// Build builds the Event.
func (rb *EventBuilder) Build() (*Event, error) {
	if rb.err != nil {
		return nil, rb.err
	}

	if _, err := rb.event.Serialize(serializer.DeSeriModePerformValidation, nil); err != nil {
		return nil, fmt.Errorf("unable to build participation: %w", err)
	}

	return rb.event, nil
}
