package dto

type InstructionsRequest struct {
	Instructions string `json:"instructions" validate:"max=1000"`
}

type AddFactRequest struct {
	Text string `json:"text" validate:"required,max=200"`
}

type SettingsResponse struct {
	Instructions string `json:"instructions"`
}

type FactResponse struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Pinned    bool   `json:"pinned"`
	CreatedAt string `json:"created_at"`
}
