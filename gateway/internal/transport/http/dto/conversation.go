package dto

type CreateConversationRequest struct {
	Title string `json:"title" validate:"omitempty,max=200"`
	Model string `json:"model" validate:"required,max=64"`
}

type RenameConversationRequest struct {
	Title string `json:"title" validate:"required,max=200"`
}

type SwitchBranchRequest struct {
	MessageID string `json:"message_id" validate:"required,uuid4"`
}

type ConversationResponse struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	DefaultModel  string  `json:"default_model"`
	HeadMessageID *string `json:"head_message_id"`
	Archived      bool    `json:"archived"`
	CreatedAt     string  `json:"created_at"`
}

type MessageResponse struct {
	ID        string   `json:"id"`
	ParentID  *string  `json:"parent_id"`
	Role      string   `json:"role"`
	Content   string   `json:"content"`
	Status    string   `json:"status"`
	Model     string   `json:"model"`
	TokensIn  int      `json:"tokens_in"`
	TokensOut int      `json:"tokens_out"`
	Error     *string  `json:"error,omitempty"`
	Siblings  []string `json:"siblings"`
	CreatedAt string   `json:"created_at"`
}

type ThreadResponse struct {
	Messages []MessageResponse `json:"messages"`
}
