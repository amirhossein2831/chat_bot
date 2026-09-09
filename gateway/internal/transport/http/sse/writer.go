package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"
)

const writeDeadline = 30 * time.Second

type metaFrame struct {
	MessageID string `json:"message_id"`
	Truncated bool   `json:"truncated"`
}

type textFrame struct {
	Text string `json:"text"`
}

type doneFrame struct {
	TokensIn  int `json:"tokens_in"`
	TokensOut int `json:"tokens_out"`
}

type errorFrame struct {
	Code string `json:"code"`
}

// Prepare sets the headers a streaming response needs and flushes them, so the
// client sees a response before the first token arrives.
func Prepare(w http.ResponseWriter) *http.ResponseController {
	h := w.Header()
	h.Set("Content-Type", "text/event-stream; charset=utf-8")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	// Without this a reverse proxy buffers the whole stream and delivers it in
	// one burst at the end, which looks exactly like a frozen product.
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	rc := http.NewResponseController(w)
	_ = rc.Flush()
	return rc
}

func Meta(w http.ResponseWriter, rc *http.ResponseController, messageID string, truncated bool) {
	write(w, rc, "meta", metaFrame{MessageID: messageID, Truncated: truncated})
}

// Pipe relays deltas until the channel closes or the client goes away. It never
// cancels the producer: a dropped connection must not destroy a generation that
// has already been paid for.
func Pipe(w http.ResponseWriter, r *http.Request, rc *http.ResponseController, deltas <-chan provider.Delta) {
	for {
		select {
		case <-r.Context().Done():
			return
		case delta, ok := <-deltas:
			if !ok {
				return
			}
			if !emit(w, rc, delta) {
				return
			}
		}
	}
}

func emit(w http.ResponseWriter, rc *http.ResponseController, delta provider.Delta) bool {
	switch delta.Type {
	case provider.DeltaText:
		return write(w, rc, "delta", textFrame{Text: delta.Text})
	case provider.DeltaDone:
		frame := doneFrame{}
		if delta.Usage != nil {
			frame.TokensIn, frame.TokensOut = delta.Usage.In, delta.Usage.Out
		}
		return write(w, rc, "done", frame)
	case provider.DeltaError:
		return write(w, rc, "error", errorFrame{Code: "upstream_error"})
	}
	return true
}

func write(w http.ResponseWriter, rc *http.ResponseController, event string, payload any) bool {
	body, err := json.Marshal(payload)
	if err != nil {
		return false
	}

	// A per-write deadline instead of Server.WriteTimeout, which would bound
	// the whole stream and truncate long answers.
	_ = rc.SetWriteDeadline(time.Now().Add(writeDeadline))

	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, body); err != nil {
		return false
	}
	_ = rc.Flush()
	return true
}
