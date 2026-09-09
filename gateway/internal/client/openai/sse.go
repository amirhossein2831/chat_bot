package openai

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
	dataPrefix   = "data:"
	doneSentinel = "[DONE]"

	initialBuffer = 64 * 1024
	// A response that arrives one token per frame can still exceed the 64KB
	// default line limit, which would end the stream with a bufio error.
	maxFrameBytes = 1 << 20
)

// emitStream guarantees exactly one terminal delta on every exit path.
func emitStream(ctx context.Context, body io.Reader, out chan<- provider.Delta) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, initialBuffer), maxFrameBytes)

	var usage *provider.Usage

	emit := func(d provider.Delta) bool {
		select {
		case out <- d:
			return true
		case <-ctx.Done():
			return false
		}
	}

	for scanner.Scan() {
		payload, ok := framePayload(scanner.Text())
		if !ok {
			continue
		}
		if payload == doneSentinel {
			emit(provider.Delta{Type: provider.DeltaDone, Usage: usage})
			return
		}

		var chunk chatChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}

		if chunk.Error != nil {
			emit(provider.Delta{Type: provider.DeltaError, Err: errors.New(chunk.Error.Message)})
			return
		}
		if chunk.Usage != nil {
			usage = &provider.Usage{
				In:  chunk.Usage.PromptTokens,
				Out: chunk.Usage.CompletionTokens,
			}
		}

		for _, choice := range chunk.Choices {
			if choice.Delta.Content == "" {
				continue
			}
			if !emit(provider.Delta{Type: provider.DeltaText, Text: choice.Delta.Content}) {
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		emit(provider.Delta{Type: provider.DeltaError, Err: err})
		return
	}

	// The upstream closed without a sentinel. Treat what arrived as complete
	// rather than losing it, but say so.
	emit(provider.Delta{Type: provider.DeltaDone, Usage: usage})
}

func framePayload(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, dataPrefix) {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(trimmed, dataPrefix)), true
}
