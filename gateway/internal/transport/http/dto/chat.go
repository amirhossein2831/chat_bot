package dto

type RegenerateRequest struct {
	Model string `json:"model" validate:"omitempty,max=64"`
}

type SendMessageRequest struct {
	ConversationID string   `json:"conversation_id" validate:"required,uuid4"`
	Content        string   `json:"content"         validate:"required,max=32000"`
	Model          string   `json:"model"           validate:"omitempty,max=64"`
	EditOf         string   `json:"edit_of"         validate:"omitempty,uuid4"`
	AttachmentIDs  []string `json:"attachment_ids" validate:"omitempty,max=10,dive,uuid4"`
}
