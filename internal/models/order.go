package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID        uuid.UUID `json:"-"`
	Number    string    `json:"number,omitempty"`
	UserID    uuid.UUID `json:"-"`
	Points    float32   `json:"accrual,omitempty"`
	Status    string    `json:"status,omitempty"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"uploaded_at,omitempty"`
}
