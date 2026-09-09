package provider

import "context"

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type DeltaType int

const (
	DeltaText DeltaType = iota
	DeltaDone
	DeltaError
)

type Tier string

const (
	TierLocal    Tier = "local"
	TierFrontier Tier = "frontier"
)

type Message struct {
	Role    Role
	Content string
}

// ToolDef is declared for phase two tool calling and is unused today. It lives
// here so adding it later changes no signatures.
type ToolDef struct {
	Name        string
	Description string
	Schema      map[string]any
}

type ChatRequest struct {
	Model       string
	System      string
	Messages    []Message
	MaxTokens   int
	Temperature float32
	Tools       []ToolDef
}

type Usage struct {
	In  int
	Out int
}

type Delta struct {
	Type  DeltaType
	Text  string
	Usage *Usage
	Err   error
}

type ModelInfo struct {
	ID            string
	Provider      string
	Tier          Tier
	ContextWindow int
	Available     bool
}

// Provider streams a completion. The channel yields zero or more DeltaText
// values and is terminated by exactly one DeltaDone carrying usage, or one
// DeltaError, and is then closed. Billing depends on that invariant.
type Provider interface {
	Stream(ctx context.Context, req ChatRequest) (<-chan Delta, error)
	Models() []ModelInfo
	Health(ctx context.Context) error
}
