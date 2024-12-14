package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/mikarwacki/chirpy/internal/database"
)

type responseChirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type responseUser struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func NewResponseUser(user database.User, token string) *responseUser {
	return &responseUser{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email, Token: token}
}

type requestUser struct {
	Email                        string `json:"email"`
	Password                     string `json:"password"`
	TokenExpiryDurationInSeconds int    `json:"expires_in_seconds,omitempty"`
}
