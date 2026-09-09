package anthropic

import "github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"

// defaultMaxTokens exists because the API rejects a request without it, unlike
// the OpenAI format where it is optional.
const defaultMaxTokens = 4096

type chatRequest struct {
	Model     string        `json:"model"`
	System    string        `json:"system,omitempty"`
	Messages  []chatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens"`
	Stream    bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type event struct {
	Type    string `json:"type"`
	Message *struct {
		Usage *struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message"`
	Delta *struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
	Usage *struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// toWire moves the system prompt to its own top-level field: this API does not
// accept a system role inside the message list.
func toWire(req provider.ChatRequest) chatRequest {
	messages := make([]chatMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		messages = append(messages, chatMessage{Role: string(m.Role), Content: m.Content})
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}

	return chatRequest{
		Model:     req.Model,
		System:    req.System,
		Messages:  messages,
		MaxTokens: maxTokens,
		Stream:    true,
	}
}
