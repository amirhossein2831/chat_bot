package dto

type PromptResponse struct {
	ID       string `json:"id"`
	Locale   string `json:"locale"`
	Category string `json:"category"`
	Title    string `json:"title"`
	Body     string `json:"body"`
}
