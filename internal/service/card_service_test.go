package service_test

import (
	"testing"

	"github.com/dewadityasanjaya/card-authorization-service/internal/dto"
	apperrors "github.com/dewadityasanjaya/card-authorization-service/internal/errors"
	"github.com/dewadityasanjaya/card-authorization-service/internal/model"
	"github.com/dewadityasanjaya/card-authorization-service/internal/repository/mock"
	"github.com/dewadityasanjaya/card-authorization-service/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	mocklib "github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ── Create Card ──────────────────────────────────────────────────

func TestCreateCard(t *testing.T) {
	tests := []struct {
		name        string
		req         dto.CreateCardRequest
		mockSetup   func(*mock.CardRepositoryMock)
		expectError bool
		errorCode   string
	}{
		{
			name: "success - create card with balance",
			req: dto.CreateCardRequest{
				CardholderName: "John Doe",
				Currency:       "SGD",
				InitialBalance: 1000.00,
			},
			mockSetup: func(m *mock.CardRepositoryMock) {
				m.On("Create", mocklib.AnythingOfType("*model.Card")).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "success - create card with zero balance",
			req: dto.CreateCardRequest{
				CardholderName: "Jane Doe",
				Currency:       "SGD",
				InitialBalance: 0,
			},
			mockSetup: func(m *mock.CardRepositoryMock) {
				m.On("Create", mocklib.AnythingOfType("*model.Card")).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "fail - database error",
			req: dto.CreateCardRequest{
				CardholderName: "John Doe",
				Currency:       "SGD",
				InitialBalance: 100,
			},
			mockSetup: func(m *mock.CardRepositoryMock) {
				m.On("Create", mocklib.AnythingOfType("*model.Card")).
					Return(assert.AnError)
			},
			expectError: true,
			errorCode:   apperrors.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cardRepo := new(mock.CardRepositoryMock)
			tt.mockSetup(cardRepo)

			svc := service.NewCardService(cardRepo)
			resp, err := svc.CreateCard(tt.req)

			if tt.expectError {
				assert.Nil(t, resp)
				assert.NotNil(t, err)
				appErr, ok := apperrors.IsAppError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, appErr.Code)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "ACTIVE", resp.Status)
				assert.Equal(t, tt.req.InitialBalance, resp.Balance)
				assert.NotEmpty(t, resp.CardNumber)
				assert.NotEmpty(t, resp.CardID)
			}

			cardRepo.AssertExpectations(t)
		})
	}
}

// ── Get Card ─────────────────────────────────────────────────────

func TestGetCard(t *testing.T) {
	cardID := uuid.New()

	tests := []struct {
		name        string
		id          string
		mockSetup   func(*mock.CardRepositoryMock)
		expectError bool
		errorCode   string
	}{
		{
			name: "success - get existing card",
			id:   cardID.String(),
			mockSetup: func(m *mock.CardRepositoryMock) {
				m.On("FindByID", cardID).Return(&model.Card{
					ID:             cardID,
					CardholderName: "John Doe",
					CardNumber:     "4556123412341234",
					Status:         model.CardStatusActive,
					Currency:       "SGD",
					Balance:        1000.00,
				}, nil)
			},
			expectError: false,
		},
		{
			name: "fail - card not found",
			id:   cardID.String(),
			mockSetup: func(m *mock.CardRepositoryMock) {
				m.On("FindByID", cardID).
					Return(nil, gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorCode:   apperrors.CodeCardNotFound,
		},
		{
			name:        "fail - invalid UUID",
			id:          "not-a-valid-uuid",
			mockSetup:   func(m *mock.CardRepositoryMock) {},
			expectError: true,
			errorCode:   apperrors.CodeCardNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cardRepo := new(mock.CardRepositoryMock)
			tt.mockSetup(cardRepo)

			svc := service.NewCardService(cardRepo)
			resp, err := svc.GetCard(tt.id)

			if tt.expectError {
				assert.Nil(t, resp)
				assert.NotNil(t, err)
				appErr, ok := apperrors.IsAppError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, appErr.Code)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, cardID.String(), resp.ID)
			}

			cardRepo.AssertExpectations(t)
		})
	}
}

// ── Freeze Card ──────────────────────────────────────────────────

func TestFreezeCard(t *testing.T) {
	cardID := uuid.New()

	tests := []struct {
		name        string
		id          string
		mockSetup   func(*mock.CardRepositoryMock)
		expectError bool
		errorCode   string
	}{
		{
			name: "success - freeze active card",
			id:   cardID.String(),
			mockSetup: func(m *mock.CardRepositoryMock) {
				m.On("FindByID", cardID).Return(&model.Card{
					ID:     cardID,
					Status: model.CardStatusActive,
				}, nil)
				m.On("Update", mocklib.AnythingOfType("*model.Card")).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "fail - card already frozen",
			id:   cardID.String(),
			mockSetup: func(m *mock.CardRepositoryMock) {
				m.On("FindByID", cardID).Return(&model.Card{
					ID:     cardID,
					Status: model.CardStatusFrozen,
				}, nil)
			},
			expectError: true,
			errorCode:   apperrors.CodeCannotFreeze,
		},
		{
			name: "fail - card not found",
			id:   cardID.String(),
			mockSetup: func(m *mock.CardRepositoryMock) {
				m.On("FindByID", cardID).
					Return(nil, gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorCode:   apperrors.CodeCardNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cardRepo := new(mock.CardRepositoryMock)
			tt.mockSetup(cardRepo)

			svc := service.NewCardService(cardRepo)
			err := svc.FreezeCard(tt.id)

			if tt.expectError {
				assert.NotNil(t, err)
				appErr, ok := apperrors.IsAppError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, appErr.Code)
			} else {
				assert.Nil(t, err)
			}

			cardRepo.AssertExpectations(t)
		})
	}
}

// ── Unfreeze Card ────────────────────────────────────────────────

func TestUnfreezeCard(t *testing.T) {
	cardID := uuid.New()

	tests := []struct {
		name        string
		id          string
		mockSetup   func(*mock.CardRepositoryMock)
		expectError bool
		errorCode   string
	}{
		{
			name: "success - unfreeze frozen card",
			id:   cardID.String(),
			mockSetup: func(m *mock.CardRepositoryMock) {
				m.On("FindByID", cardID).Return(&model.Card{
					ID:     cardID,
					Status: model.CardStatusFrozen,
				}, nil)
				m.On("Update", mocklib.AnythingOfType("*model.Card")).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "fail - card already active",
			id:   cardID.String(),
			mockSetup: func(m *mock.CardRepositoryMock) {
				m.On("FindByID", cardID).Return(&model.Card{
					ID:     cardID,
					Status: model.CardStatusActive,
				}, nil)
			},
			expectError: true,
			errorCode:   apperrors.CodeCannotUnfreeze,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cardRepo := new(mock.CardRepositoryMock)
			tt.mockSetup(cardRepo)

			svc := service.NewCardService(cardRepo)
			err := svc.UnfreezeCard(tt.id)

			if tt.expectError {
				assert.NotNil(t, err)
				appErr, ok := apperrors.IsAppError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, appErr.Code)
			} else {
				assert.Nil(t, err)
			}

			cardRepo.AssertExpectations(t)
		})
	}
}

// ── Top Up ───────────────────────────────────────────────────────

func TestTopUp(t *testing.T) {
	cardID := uuid.New()

	tests := []struct {
		name            string
		id              string
		req             dto.TopUpRequest
		mockSetup       func(*mock.CardRepositoryMock)
		expectError     bool
		errorCode       string
		expectedBalance float64
	}{
		{
			name: "success - top up active card",
			id:   cardID.String(),
			req:  dto.TopUpRequest{Amount: 500},
			mockSetup: func(m *mock.CardRepositoryMock) {
				m.On("FindByID", cardID).Return(&model.Card{
					ID:       cardID,
					Status:   model.CardStatusActive,
					Balance:  1000.00,
					Currency: "SGD",
				}, nil)
				m.On("Update", mocklib.AnythingOfType("*model.Card")).
					Return(nil)
			},
			expectError:     false,
			expectedBalance: 1500.00,
		},
		{
			name: "fail - top up frozen card",
			id:   cardID.String(),
			req:  dto.TopUpRequest{Amount: 500},
			mockSetup: func(m *mock.CardRepositoryMock) {
				m.On("FindByID", cardID).Return(&model.Card{
					ID:     cardID,
					Status: model.CardStatusFrozen,
				}, nil)
			},
			expectError: true,
			errorCode:   apperrors.CodeCardFrozen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cardRepo := new(mock.CardRepositoryMock)
			tt.mockSetup(cardRepo)

			svc := service.NewCardService(cardRepo)
			resp, err := svc.TopUp(tt.id, tt.req)

			if tt.expectError {
				assert.Nil(t, resp)
				assert.NotNil(t, err)
				appErr, ok := apperrors.IsAppError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, appErr.Code)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedBalance, resp.Balance)
			}

			cardRepo.AssertExpectations(t)
		})
	}
}
