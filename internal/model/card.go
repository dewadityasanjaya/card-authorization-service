package model

import (
	"time"

	"github.com/google/uuid"
)

type CardStatus string

const (
	CardStatusActive CardStatus = "ACTIVE"
	CardStatusFrozen CardStatus = "FROZEN"
)

type Card struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CardNumber     string     `gorm:"type:varchar(16);not null;unique"`
	CardholderName string     `gorm:"type:varchar(255);not null"`
	Status         CardStatus `gorm:"type:card_status;not null;default:'ACTIVE'"`
	Currency       string     `gorm:"type:varchar(3);not null"`
	Balance        float64    `gorm:"type:numeric(19,4);not null;default:0"`
	CreatedAt      time.Time  `gorm:"not null;default:now()"`
	UpdatedAt      time.Time  `gorm:"not null;default:now()"`
}

func (Card) TableName() string {
	return "cards"
}
