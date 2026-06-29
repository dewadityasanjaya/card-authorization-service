package mock

import (
	"github.com/dewadityasanjaya/card-authorization-service/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type CardRepositoryMock struct {
	mock.Mock
}

func (m *CardRepositoryMock) Create(card *model.Card) error {
	args := m.Called(card)
	return args.Error(0)
}

func (m *CardRepositoryMock) FindByID(id uuid.UUID) (*model.Card, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Card), args.Error(1)
}

func (m *CardRepositoryMock) FindByCardNumber(cardNumber string) (*model.Card, error) {
	args := m.Called(cardNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Card), args.Error(1)
}

func (m *CardRepositoryMock) Update(card *model.Card) error {
	args := m.Called(card)
	return args.Error(0)
}

func (m *CardRepositoryMock) FindByIDForUpdate(tx *gorm.DB, id uuid.UUID) (*model.Card, error) {
	args := m.Called(tx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Card), args.Error(1)
}
