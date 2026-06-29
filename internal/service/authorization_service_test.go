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

// ── Authorize ────────────────────────────────────────────────────

func TestAuthorize(t *testing.T) {
	cardID := uuid.New()
	cardNumber := "4556123412341234"

	baseCard := &model.Card{
		ID:         cardID,
		CardNumber: cardNumber,
		Status:     model.CardStatusActive,
		Currency:   "SGD",
		Balance:    1000.00,
	}

	baseReq := dto.AuthorizeRequest{
		CardNumber:   cardNumber,
		MerchantID:   "STARBUCKS001",
		MerchantName: "Starbucks",
		Currency:     "SGD",
		Amount:       25.50,
	}

	tests := []struct {
		name           string
		req            dto.AuthorizeRequest
		idempotencyKey string
		mockSetup      func(*mock.CardRepositoryMock, *mock.AuthorizationRepositoryMock)
		expectApproved bool
		expectDeclined bool
		declineReason  string
		remainingBal   float64
	}{
		{
			name:           "success - approved authorization",
			req:            baseReq,
			idempotencyKey: "",
			mockSetup: func(c *mock.CardRepositoryMock, a *mock.AuthorizationRepositoryMock) {
				c.On("FindByCardNumber", cardNumber).Return(baseCard, nil)
				c.On("FindByIDForUpdate", mocklib.Anything, cardID).Return(baseCard, nil)
				c.On("Update", mocklib.AnythingOfType("*model.Card")).Return(nil)
				a.On("Create", mocklib.Anything, mocklib.AnythingOfType("*model.Authorization")).Return(nil)
			},
			expectApproved: true,
			remainingBal:   974.50,
		},
		{
			name:           "declined - card not found",
			req:            baseReq,
			idempotencyKey: "",
			mockSetup: func(c *mock.CardRepositoryMock, a *mock.AuthorizationRepositoryMock) {
				c.On("FindByCardNumber", cardNumber).
					Return(nil, gorm.ErrRecordNotFound)
			},
			expectDeclined: true,
			declineReason:  apperrors.CodeCardNotFound,
		},
		{
			name: "declined - card frozen",
			req:  baseReq,
			mockSetup: func(c *mock.CardRepositoryMock, a *mock.AuthorizationRepositoryMock) {
				frozenCard := *baseCard
				frozenCard.Status = model.CardStatusFrozen
				c.On("FindByCardNumber", cardNumber).Return(&frozenCard, nil)
			},
			expectDeclined: true,
			declineReason:  apperrors.CodeCardFrozen,
		},
		{
			name: "declined - currency mismatch",
			req: dto.AuthorizeRequest{
				CardNumber:   cardNumber,
				MerchantID:   "STARBUCKS001",
				MerchantName: "Starbucks",
				Currency:     "USD", // card is SGD
				Amount:       25.50,
			},
			mockSetup: func(c *mock.CardRepositoryMock, a *mock.AuthorizationRepositoryMock) {
				c.On("FindByCardNumber", cardNumber).Return(baseCard, nil)
			},
			expectDeclined: true,
			declineReason:  apperrors.CodeInvalidCurrency,
		},
		{
			name: "declined - insufficient funds",
			req: dto.AuthorizeRequest{
				CardNumber:   cardNumber,
				MerchantID:   "STARBUCKS001",
				MerchantName: "Starbucks",
				Currency:     "SGD",
				Amount:       9999.00, // more than balance
			},
			mockSetup: func(c *mock.CardRepositoryMock, a *mock.AuthorizationRepositoryMock) {
				c.On("FindByCardNumber", cardNumber).Return(baseCard, nil)
				c.On("FindByIDForUpdate", mocklib.Anything, cardID).Return(baseCard, nil)
			},
			expectDeclined: true,
			declineReason:  apperrors.CodeInsufficientFunds,
		},
		{
			name:           "success - with idempotency key",
			req:            baseReq,
			idempotencyKey: "my-idempotency-key",
			mockSetup: func(c *mock.CardRepositoryMock, a *mock.AuthorizationRepositoryMock) {
				a.On("FindIdempotencyKey", "my-idempotency-key").
					Return(nil, gorm.ErrRecordNotFound)
				c.On("FindByCardNumber", cardNumber).Return(baseCard, nil)
				c.On("FindByIDForUpdate", mocklib.Anything, cardID).Return(baseCard, nil)
				c.On("Update", mocklib.AnythingOfType("*model.Card")).Return(nil)
				a.On("Create", mocklib.Anything, mocklib.AnythingOfType("*model.Authorization")).Return(nil)
				a.On("SaveIdempotencyKey", mocklib.Anything, mocklib.AnythingOfType("*model.IdempotencyKey")).Return(nil)
			},
			expectApproved: true,
			remainingBal:   949,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cardRepo := new(mock.CardRepositoryMock)
			authRepo := new(mock.AuthorizationRepositoryMock)
			tt.mockSetup(cardRepo, authRepo)

			// Use a real gorm.DB with SQLite or nil for unit tests
			txManager := new(mock.MockTxManager)
			svc := service.NewAuthorizationService(txManager, cardRepo, authRepo)
			approved, declined, err := svc.Authorize(tt.req, tt.idempotencyKey)

			assert.Nil(t, err)

			if tt.expectApproved {
				assert.NotNil(t, approved)
				assert.Nil(t, declined)
				assert.Equal(t, "APPROVED", approved.Status)
				assert.Equal(t, tt.remainingBal, approved.RemainingBalance)
			}

			if tt.expectDeclined {
				assert.Nil(t, approved)
				assert.NotNil(t, declined)
				assert.Equal(t, "DECLINED", declined.Status)
				assert.Equal(t, tt.declineReason, declined.Reason)
			}

			cardRepo.AssertExpectations(t)
			authRepo.AssertExpectations(t)
		})
	}
}

// ── Reverse ──────────────────────────────────────────────────────

func TestReverse(t *testing.T) {
	cardID := uuid.New()
	authID := uuid.New()
	authCode := "AUTH-ABC12345"

	tests := []struct {
		name            string
		authorizationID string
		mockSetup       func(*mock.CardRepositoryMock, *mock.AuthorizationRepositoryMock)
		expectError     bool
		errorCode       string
	}{
		{
			name:            "success - reverse approved authorization",
			authorizationID: authCode,
			mockSetup: func(c *mock.CardRepositoryMock, a *mock.AuthorizationRepositoryMock) {
				a.On("FindByAuthorizationCode", authCode).Return(&model.Authorization{
					ID:                authID,
					AuthorizationCode: authCode,
					CardID:            cardID,
					Amount:            25.50,
					Status:            model.AuthorizationStatusApproved,
				}, nil)
				c.On("FindByIDForUpdate", mocklib.Anything, cardID).Return(&model.Card{
					ID:      cardID,
					Balance: 974.50,
				}, nil)
				c.On("Update", mocklib.AnythingOfType("*model.Card")).Return(nil)
				a.On("Update", mocklib.Anything, mocklib.AnythingOfType("*model.Authorization")).Return(nil)
			},
			expectError: false,
		},
		{
			name:            "fail - already reversed",
			authorizationID: authCode,
			mockSetup: func(c *mock.CardRepositoryMock, a *mock.AuthorizationRepositoryMock) {
				a.On("FindByAuthorizationCode", authCode).Return(&model.Authorization{
					ID:                authID,
					AuthorizationCode: authCode,
					CardID:            cardID,
					Amount:            25.50,
					Status:            model.AuthorizationStatusReversed,
				}, nil)
			},
			expectError: true,
			errorCode:   apperrors.CodeAuthAlreadyReversed,
		},
		{
			name:            "fail - authorization not found",
			authorizationID: authCode,
			mockSetup: func(c *mock.CardRepositoryMock, a *mock.AuthorizationRepositoryMock) {
				a.On("FindByAuthorizationCode", authCode).
					Return(nil, gorm.ErrRecordNotFound)
			},
			expectError: true,
			errorCode:   apperrors.CodeAuthNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cardRepo := new(mock.CardRepositoryMock)
			authRepo := new(mock.AuthorizationRepositoryMock)
			tt.mockSetup(cardRepo, authRepo)

			txManager := new(mock.MockTxManager)
			svc := service.NewAuthorizationService(txManager, cardRepo, authRepo)
			resp, err := svc.Reverse(tt.authorizationID)

			if tt.expectError {
				assert.Nil(t, resp)
				assert.NotNil(t, err)
				appErr, ok := apperrors.IsAppError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, appErr.Code)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "REVERSED", resp.Status)
				assert.Equal(t, 25.50, resp.RefundedAmount)
			}

			cardRepo.AssertExpectations(t)
			authRepo.AssertExpectations(t)
		})
	}
}
