package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
	"github.com/amirhossein2831/chat_bot/gateway/internal/util/password"
)

type Service struct {
	users    Users
	sessions Sessions
}

func NewService(users Users, sessions Sessions) *Service {
	return &Service{users: users, sessions: sessions}
}

func (s *Service) Register(ctx context.Context, user model.User, plain, invite string) (model.User, error) {
	hash, err := password.Hash(plain)
	if err != nil {
		return model.User{}, fmt.Errorf("hashing password: %w", err)
	}

	user.ID = uuid.New()
	user.PasswordHash = hash
	user.Status = model.UserStatusActive
	if user.Locale == "" {
		user.Locale = "fa"
	}
	if user.DirPref == "" {
		user.DirPref = "auto"
	}

	if err := s.users.Create(ctx, user, invite); err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, email, plain string) (string, model.User, error) {
	user, err := s.users.ByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			// Indistinguishable from a wrong password on purpose: otherwise
			// this endpoint enumerates registered addresses.
			return "", model.User{}, constants.ErrUnauthorized
		}
		return "", model.User{}, err
	}

	ok, err := password.Verify(plain, user.PasswordHash)
	if err != nil || !ok {
		return "", model.User{}, constants.ErrUnauthorized
	}
	if !user.IsActive() {
		return "", model.User{}, constants.ErrForbidden
	}

	token, err := s.sessions.Create(ctx, user.ID)
	if err != nil {
		return "", model.User{}, err
	}
	return token, user, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (model.User, error) {
	id, err := s.sessions.UserID(ctx, token)
	if err != nil {
		return model.User{}, constants.ErrUnauthorized
	}

	user, err := s.users.ByID(ctx, id)
	if err != nil {
		return model.User{}, constants.ErrUnauthorized
	}
	if !user.IsActive() {
		return model.User{}, constants.ErrForbidden
	}
	return user, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.sessions.Delete(ctx, token)
}
