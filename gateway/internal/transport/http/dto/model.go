package dto

type ModelResponse struct {
	ID            string `json:"id"`
	Provider      string `json:"provider"`
	Tier          string `json:"tier"`
	ContextWindow int    `json:"context_window"`
	Available     bool   `json:"available"`
}
