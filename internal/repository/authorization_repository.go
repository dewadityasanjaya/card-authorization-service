package repository

import (
	"github.com/dewadityasanjaya/card-authorization-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthorizationRepository interface {
	Create(tx *gorm.DB, auth *model.Authorization) error
	FindByAuthorizationCode(code string) (*model.Authorization, error)
	FindByCardID(cardID uuid.UUID) ([]model.Authorization, error)
	Update(tx *gorm.DB, auth *model.Authorization) error
	FindIdempotencyKey(key string) (*model.IdempotencyKey, error)
	SaveIdempotencyKey(tx *gorm.DB, idempotencyKey *model.IdempotencyKey) error
}

type authorizationRepository struct {
	db *gorm.DB
}

func NewAuthorizationRepository(db *gorm.DB) AuthorizationRepository {
	return &authorizationRepository{db: db}
}

// Create inserts a new authorization inside a transaction
func (r *authorizationRepository) Create(tx *gorm.DB, auth *model.Authorization) error {
	return tx.Create(auth).Error
}

// FindByAuthorizationCode finds an authorization by its code (e.g. AUTH-XXXX)
func (r *authorizationRepository) FindByAuthorizationCode(code string) (*model.Authorization, error) {
	var auth model.Authorization
	err := r.db.Where("authorization_code = ?", code).First(&auth).Error
	if err != nil {
		return nil, err
	}
	return &auth, nil
}

// FindByCardID returns all authorizations for a card, newest first
func (r *authorizationRepository) FindByCardID(cardID uuid.UUID) ([]model.Authorization, error) {
	var auths []model.Authorization
	err := r.db.Where("card_id = ?", cardID).
		Order("created_at DESC").
		Find(&auths).Error
	if err != nil {
		return nil, err
	}
	return auths, nil
}

// Update saves changes to an existing authorization inside a transaction
func (r *authorizationRepository) Update(tx *gorm.DB, auth *model.Authorization) error {
	return tx.Save(auth).Error
}

// FindIdempotencyKey checks if an idempotency key already exists
func (r *authorizationRepository) FindIdempotencyKey(key string) (*model.IdempotencyKey, error) {
	var idempotencyKey model.IdempotencyKey
	err := r.db.Where("key = ?", key).First(&idempotencyKey).Error
	if err != nil {
		return nil, err
	}
	return &idempotencyKey, nil
}

// SaveIdempotencyKey stores a new idempotency key inside a transaction
func (r *authorizationRepository) SaveIdempotencyKey(tx *gorm.DB, idempotencyKey *model.IdempotencyKey) error {
	return tx.Create(idempotencyKey).Error
}
