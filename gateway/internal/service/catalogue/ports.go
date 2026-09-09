package catalogue

import "github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"

type Registry interface {
	Models() []provider.ModelInfo
	ContextWindow(modelID string) int
}
