package anthropic

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"
)

const (
	dataPrefix    = "data:"
	initialBuffer = 64 * 1024
	maxFrameBytes = 1 << 20

	eventContentBlockDelta = "content_block_delta"
	eventMessageStart      = "message_start"
	eventMessageDelta      = "message_delta"
	eventMessageStop       = "message_stop"
	eventError             = "error"
	deltaText              = "text_delta"
)

// emitStream guarantees exactly one terminal delta on every exit path. Usage
// arrives split across two event types: input tokens on message_start, output
// tokens on message_delta.
func emitStream(ctx context.Context, body io.Reader, out chan<- provider.Delta) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, initialBuffer), maxFrameBytes)

	usage := provider.Usage{}
	sawUsage := false

	emit := func(d provider.Delta) bool {
		select {
		case out <- d:
			return true
		case <-ctx.Done():
			return false
		}
	}
	terminate := func() {
		if sawUsage {
			emit(provider.Delta{Type: provider.DeltaDone, Usage: &usage})
			return
		}
		emit(provider.Delta{Type: provider.DeltaDone})
	}

	for scanner.Scan() {
		payload, ok := framePayload(scanner.Text())
		if !ok {
			continue
		}

		var ev event
		if err := json.Unmarshal([]byte(payload), &ev); err != nil {
			continue
		}

		switch ev.Type {
		case eventError:
			message := "upstream error"
			if ev.Error != nil {
				message = ev.Error.Message
			}
			emit(provider.Delta{Type: provider.DeltaError, Err: errors.New(message)})
			return

		case eventMessageStart:
			if ev.Message != nil && ev.Message.Usage != nil {
				usage.In = ev.Message.Usage.InputTokens
				sawUsage = true
			}

		case eventMessageDelta:
			if ev.Usage != nil {
				usage.Out = ev.Usage.OutputTokens
				sawUsage = true
			}

		case eventContentBlockDelta:
			if ev.Delta == nil || ev.Delta.Type != deltaText || ev.Delta.Text == "" {
				continue
			}
			if !emit(provider.Delta{Type: provider.DeltaText, Text: ev.Delta.Text}) {
				return
			}

		case eventMessageStop:
			terminate()
			return
		}
	}

	if err := scanner.Err(); err != nil {
		emit(provider.Delta{Type: provider.DeltaError, Err: err})
		return
	}
	terminate()
}

func framePayload(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, dataPrefix) {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(trimmed, dataPrefix)), true
}
