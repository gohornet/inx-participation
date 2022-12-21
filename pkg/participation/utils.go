package participation

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
)

//nolint:revive // better be explicit here
type ParticipationBlock struct {
	BlockID iotago.BlockID
	Block   *iotago.Block
	Data    []byte
}

//nolint:revive // better be explicit here
type ParticipationOutput struct {
	BlockID  iotago.BlockID
	OutputID iotago.OutputID
	Address  iotago.Address
	Deposit  uint64
}

func (o *ParticipationOutput) serializedAddressBytes() ([]byte, error) {
	return o.Address.Serialize(serializer.DeSeriModeNoValidation, nil)
}

func (b *ParticipationBlock) Transaction() *iotago.Transaction {
	switch payload := b.Block.Payload.(type) {
	case *iotago.Transaction:
		return payload
	default:
		return nil
	}
}

func (b *ParticipationBlock) TransactionEssence() *iotago.TransactionEssence {
	if transaction := b.Transaction(); transaction != nil {
		return transaction.Essence
	}

	return nil
}

func (b *ParticipationBlock) TransactionEssenceTaggedData() *iotago.TaggedData {
	if essence := b.TransactionEssence(); essence != nil {
		switch payload := essence.Payload.(type) {
		case *iotago.TaggedData:
			return payload
		default:
			return nil
		}
	}

	return nil
}

func (b *ParticipationBlock) TransactionEssenceUTXOInputs() iotago.OutputIDs {
	var inputs iotago.OutputIDs
	if essence := b.TransactionEssence(); essence != nil {
		for _, input := range essence.Inputs {
			switch utxoInput := input.(type) {
			case *iotago.UTXOInput:
				id := utxoInput.ID()
				inputs = append(inputs, id)
			default:
				return nil
			}
		}
	}

	return inputs
}
