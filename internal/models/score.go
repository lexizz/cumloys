package models

import (
	"time"

	"github.com/google/uuid"
)

type Score struct {
	ID        uuid.UUID `json:"id,omitempty"`
	Total     float32   `json:"total,omitempty"`
	UserID    uuid.UUID `json:"userId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TotalScoreWithdraw struct {
	Total    float32 `json:"current"`
	Withdraw float32 `json:"withdrawn"`
}

type ScoreWithdraw struct {
	NumberOrder string    `json:"order,omitempty"`
	SumWithdraw float32   `json:"sum,omitempty"`
	CreatedAt   time.Time `json:"processed_at,omitempty"`
}
