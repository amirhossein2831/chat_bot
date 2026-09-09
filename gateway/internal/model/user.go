package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive = "active"
	StatusBanned = "banned"
)

type User struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Locale       string    `db:"locale"`
	DirPref      string    `db:"dir_pref"`
	Status       string    `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
}

func (u User) IsActive() bool { return u.Status == StatusActive }
