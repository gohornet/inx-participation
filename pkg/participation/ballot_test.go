//nolint:scopelint // we don't care about these linters in test cases
package participation_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/marshalutil"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/inx-participation/pkg/participation"
)

func RandBallot(questionCount int) (*participation.Ballot, []byte) {

	b := &participation.Ballot{
		Questions: participation.Questions{},
	}

	var questionsBytes [][]byte
	for i := 0; i < questionCount; i++ {
		q, bytes := RandValidQuestion()
		b.Questions = append(b.Questions, q)
		questionsBytes = append(questionsBytes, bytes)
	}

	ms := marshalutil.New()
	ms.WriteUint32(participation.BallotPayloadTypeID)
	ms.WriteUint8(uint8(len(questionsBytes)))
	for _, bytes := range questionsBytes {
		ms.WriteBytes(bytes)
	}

	return b, ms.Bytes()
}

func TestBallot_Deserialize(t *testing.T) {
	validBallot, validBallotData := RandBallot(1)
	maxQuestionsBallot, maxQuestionsBallotData := RandBallot(10)
	noQuestions, noQuestionsBallotData := RandBallot(0)
	tooManyQuestionsBallot, tooManyQuestionsBallotData := RandBallot(11)

	tests := []struct {
		name   string
		data   []byte
		target *participation.Ballot
		err    error
	}{
		{"ok", validBallotData, validBallot, nil},
		{"not enough data", validBallotData[:len(validBallotData)-1], validBallot, serializer.ErrDeserializationNotEnoughData},
		{"max questions", maxQuestionsBallotData, maxQuestionsBallot, nil},
		{"no questions", noQuestionsBallotData, noQuestions, serializer.ErrArrayValidationMinElementsNotReached},
		{"too many questions", tooManyQuestionsBallotData, tooManyQuestionsBallot, serializer.ErrArrayValidationMaxElementsExceeded},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &participation.Ballot{}
			bytesRead, err := u.Deserialize(tt.data, serializer.DeSeriModePerformValidation, nil)
			if tt.err != nil {
				require.True(t, errors.Is(err, tt.err))

				return
			}
			require.Equal(t, len(tt.data), bytesRead)
			require.EqualValues(t, tt.target, u)
		})
	}
}

func TestBallot_Serialize(t *testing.T) {
	validBallot, validBallotData := RandBallot(1)
	maxQuestionsBallot, maxQuestionsBallotData := RandBallot(10)
	noQuestions, noQuestionsBallotData := RandBallot(0)
	tooManyQuestionsBallot, tooManyQuestionsBallotData := RandBallot(11)

	tests := []struct {
		name   string
		source *participation.Ballot
		target []byte
		err    error
	}{
		{"ok", validBallot, validBallotData, nil},
		{"max questions", maxQuestionsBallot, maxQuestionsBallotData, nil},
		{"no questions", noQuestions, noQuestionsBallotData, serializer.ErrArrayValidationMinElementsNotReached},
		{"too many questions", tooManyQuestionsBallot, tooManyQuestionsBallotData, serializer.ErrArrayValidationMaxElementsExceeded},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.source.Serialize(serializer.DeSeriModePerformValidation, nil)
			if tt.err != nil {
				require.True(t, errors.Is(err, tt.err))

				return
			}
			require.EqualValues(t, tt.target, data)
		})
	}
}
