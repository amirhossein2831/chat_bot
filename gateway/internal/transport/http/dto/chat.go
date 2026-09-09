package dto

type SendMessageRequest struct {
	ConversationID string `json:"conversation_id" validate:"required,uuid4"`
	Content        string `json:"content"         validate:"required,max=32000"`
	Model          string `json:"model"           validate:"omitempty,max=64"`
}
