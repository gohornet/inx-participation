package participation

import (
	"context"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/inx-app/pkg/nodebridge"
	"github.com/iotaledger/inx-participation/pkg/participation"
	inx "github.com/iotaledger/inx/go"
	iotago "github.com/iotaledger/iota.go/v3"
)

func participationOutputFromINXOutput(output *inx.LedgerOutput) *participation.ParticipationOutput {
	iotaOutput, err := output.UnwrapOutput(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil
	}

	// Ignore anything other than BasicOutputs
	if iotaOutput.Type() != iotago.OutputBasic {
		return nil
	}

	unlockConditions := iotaOutput.UnlockConditionSet()

	return &participation.ParticipationOutput{
		BlockID:  output.UnwrapBlockID(),
		OutputID: output.UnwrapOutputID(),
		Address:  unlockConditions.Address().Address,
		Deposit:  iotaOutput.Deposit(),
	}
}

func NodeStatus(ctx context.Context) (confirmedIndex iotago.MilestoneIndex, pruningIndex iotago.MilestoneIndex) {
	status := deps.NodeBridge.NodeStatus()

	return status.GetConfirmedMilestone().GetMilestoneInfo().GetMilestoneIndex(), status.GetTanglePruningIndex()
}

func BlockForBlockID(ctx context.Context, blockID iotago.BlockID) (*participation.ParticipationBlock, error) {
	block, err := deps.NodeBridge.Client().ReadBlock(ctx, inx.NewBlockId(blockID))
	if err != nil {
		return nil, err
	}

	iotagoBlock, err := block.UnwrapBlock(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, err
	}

	return &participation.ParticipationBlock{
		BlockID: blockID,
		Block:   iotagoBlock,
		Data:    block.GetData(),
	}, nil
}

func OutputForOutputID(ctx context.Context, outputID iotago.OutputID) (*participation.ParticipationOutput, error) {
	resp, err := deps.NodeBridge.Client().ReadOutput(ctx, inx.NewOutputId(outputID))
	if err != nil {
		return nil, err
	}
	switch resp.GetPayload().(type) {

	//nolint:nosnakecase // grpc uses underscores
	case *inx.OutputResponse_Output:
		return participationOutputFromINXOutput(resp.GetOutput()), nil

	//nolint:nosnakecase // grpc uses underscores
	case *inx.OutputResponse_Spent:
		return participationOutputFromINXOutput(resp.GetSpent().GetOutput()), nil

	default:
		return nil, fmt.Errorf("invalid inx.OutputResponse payload type")
	}
}

func LedgerUpdates(ctx context.Context, startIndex iotago.MilestoneIndex, endIndex iotago.MilestoneIndex, handler func(index iotago.MilestoneIndex, created []*participation.ParticipationOutput, consumed []*participation.ParticipationOutput) error) error {
	return deps.NodeBridge.ListenToLedgerUpdates(ctx, startIndex, endIndex, func(update *nodebridge.LedgerUpdate) error {
		index := update.MilestoneIndex

		var created []*participation.ParticipationOutput
		for _, output := range update.Created {
			o := participationOutputFromINXOutput(output)
			if o != nil {
				created = append(created, o)
			}
		}

		var consumed []*participation.ParticipationOutput
		for _, spent := range update.Consumed {
			o := participationOutputFromINXOutput(spent.GetOutput())
			if o != nil {
				consumed = append(consumed, o)
			}
		}

		return handler(index, created, consumed)
	})
}
