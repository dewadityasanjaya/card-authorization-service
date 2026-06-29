package mock

import (
	"github.com/dewadityasanjaya/card-authorization-service/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type AuthorizationRepositoryMock struct {
	mock.Mock
}

func (m *AuthorizationRepositoryMock) Create(tx *gorm.DB, auth *model.Authorization) error {
	args := m.Called(tx, auth)
	return args.Error(0)
}

func (m *AuthorizationRepositoryMock) FindByAuthorizationCode(code string) (*model.Authorization, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Authorization), args.Error(1)
}

func (m *AuthorizationRepositoryMock) FindByCardID(cardID uuid.UUID) ([]model.Authorization, error) {
	args := m.Called(cardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Authorization), args.Error(1)
}

func (m *AuthorizationRepositoryMock) Update(tx *gorm.DB, auth *model.Authorization) error {
	args := m.Called(tx, auth)
	return args.Error(0)
}

func (m *AuthorizationRepositoryMock) FindIdempotencyKey(key string) (*model.IdempotencyKey, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.IdempotencyKey), args.Error(1)
}

func (m *AuthorizationRepositoryMock) SaveIdempotencyKey(tx *gorm.DB, idempotencyKey *model.IdempotencyKey) error {
	args := m.Called(tx, idempotencyKey)
	return args.Error(0)
}
