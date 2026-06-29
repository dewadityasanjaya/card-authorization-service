package model

import (
	"time"

	"github.com/google/uuid"
)

type AuthorizationStatus string

const (
	AuthorizationStatusApproved AuthorizationStatus = "APPROVED"
	AuthorizationStatusDeclined AuthorizationStatus = "DECLINED"
	AuthorizationStatusReversed AuthorizationStatus = "REVERSED"
)

type Authorization struct {
	ID                uuid.UUID           `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	AuthorizationCode string              `gorm:"type:varchar(20);not null;unique"`
	CardID            uuid.UUID           `gorm:"type:uuid;not null"`
	MerchantID        string              `gorm:"type:varchar(255);not null"`
	MerchantName      string              `gorm:"type:varchar(255);not null"`
	Amount            float64             `gorm:"type:numeric(19,4);not null"`
	Currency          string              `gorm:"type:varchar(3);not null"`
	Status            AuthorizationStatus `gorm:"type:authorization_status;not null"`
	CreatedAt         time.Time           `gorm:"not null;default:now()"`

	// Relationship
	Card Card `gorm:"foreignKey:CardID"`
}

func (Authorization) TableName() string {
	return "authorizations"
}
