package dto

type RegisterRequest struct {
	Email    string `json:"email"    validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,min=10,max=200"`
	Invite   string `json:"invite"   validate:"omitempty,max=64"`
	Locale   string `json:"locale"   validate:"omitempty,oneof=fa en"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,max=200"`
}

type UserResponse struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Locale  string `json:"locale"`
	DirPref string `json:"dir_pref"`
}
