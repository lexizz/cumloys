package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	IncreasePointsType int = 1
	DecreasePointsType int = 2
)

type Transaction struct {
	ID        uuid.UUID `json:"id,omitempty"`
	UserID    uuid.UUID `json:"userId,omitempty"`
	OrderID   uuid.UUID `json:"orderId,omitempty"`
	Points    float32   `json:"points,omitempty"`
	Type      int       `json:"type,omitempty"` // пополнение или списание баллов
	CreatedAt time.Time `json:"createdAt"`
}
