package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarkCompleteFromStreaming(t *testing.T) {
	m := Message{Status: StatusStreaming}

	err := m.MarkComplete(Usage{In: 100, Out: 250, CostUSDMicros: 4200})

	require.NoError(t, err)
	require.Equal(t, StatusComplete, m.Status)
	require.Equal(t, 100, m.TokensIn)
	require.Equal(t, 250, m.TokensOut)
	require.Equal(t, int64(4200), m.CostUSDMicros)
}

func TestMarkCompleteRejectsTerminalMessage(t *testing.T) {
	for _, status := range []Status{StatusComplete, StatusFailed} {
		m := Message{Status: status}
		require.ErrorIs(t, m.MarkComplete(Usage{}), ErrInvalidTransition,
			"completing a %s message must not be allowed", status)
	}
}

func TestMarkFailedRecordsReasonAndIsTerminal(t *testing.T) {
	m := Message{Status: StatusStreaming}

	m.MarkFailed("upstream timeout")

	require.Equal(t, StatusFailed, m.Status)
	require.NotNil(t, m.Error)
	require.Equal(t, "upstream timeout", *m.Error)
	require.True(t, m.IsTerminal())
}

func TestStreamingMessageIsNotTerminal(t *testing.T) {
	require.False(t, Message{Status: StatusStreaming}.IsTerminal())
}

func TestCostIsIntegerMicros(t *testing.T) {
	m := Message{Status: StatusStreaming}
	require.NoError(t, m.MarkComplete(Usage{CostUSDMicros: 1}))
	require.Equal(t, int64(1), m.CostUSDMicros)
}
