package mapper

import (
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/dto"
)

func ToUserModel(req dto.RegisterRequest) model.User {
	return model.User{
		Email:  req.Email,
		Locale: req.Locale,
	}
}

func ToUserResponse(user model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:      user.ID.String(),
		Email:   user.Email,
		Locale:  user.Locale,
		DirPref: user.DirPref,
	}
}
