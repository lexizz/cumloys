package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id,omitempty"`
	Login     string    `json:"login,omitempty"`
	Password  string    `json:"password,omitempty"`
	Token     JWT       `json:"token,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}
