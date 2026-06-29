package repository

import (
	"github.com/dewadityasanjaya/card-authorization-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CardRepository interface {
	Create(card *model.Card) error
	FindByID(id uuid.UUID) (*model.Card, error)
	FindByCardNumber(cardNumber string) (*model.Card, error)
	Update(card *model.Card) error
	FindByIDForUpdate(tx *gorm.DB, id uuid.UUID) (*model.Card, error)
}

type cardRepository struct {
	db *gorm.DB
}

func NewCardRepository(db *gorm.DB) CardRepository {
	return &cardRepository{db: db}
}

// Create inserts a new card into the database
func (r *cardRepository) Create(card *model.Card) error {
	return r.db.Create(card).Error
}

// FindByID finds a card by its UUID
func (r *cardRepository) FindByID(id uuid.UUID) (*model.Card, error) {
	var card model.Card
	err := r.db.Where("id = ?", id).First(&card).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// FindByCardNumber finds a card by its card number
func (r *cardRepository) FindByCardNumber(cardNumber string) (*model.Card, error) {
	var card model.Card
	err := r.db.Where("card_number = ?", cardNumber).First(&card).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// Update saves changes to an existing card
func (r *cardRepository) Update(card *model.Card) error {
	return r.db.Save(card).Error
}

// FindByIDForUpdate locks the row for the duration of the transaction
// This prevents race conditions during authorization (SELECT FOR UPDATE)
func (r *cardRepository) FindByIDForUpdate(tx *gorm.DB, id uuid.UUID) (*model.Card, error) {
	var card model.Card
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&card).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}
