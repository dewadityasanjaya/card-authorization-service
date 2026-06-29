package model

import (
	"time"

	"github.com/google/uuid"
)

type IdempotencyKey struct {
	Key             string    `gorm:"type:varchar(255);primary_key"`
	AuthorizationID uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt       time.Time `gorm:"not null;default:now()"`

	// Relationship
	Authorization Authorization `gorm:"foreignKey:AuthorizationID"`
}

func (IdempotencyKey) TableName() string {
	return "idempotency_keys"
}
