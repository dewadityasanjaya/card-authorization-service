package service

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/dewadityasanjaya/card-authorization-service/internal/dto"
	apperrors "github.com/dewadityasanjaya/card-authorization-service/internal/errors"
	"github.com/dewadityasanjaya/card-authorization-service/internal/model"
	"github.com/dewadityasanjaya/card-authorization-service/internal/repository"
	"github.com/dewadityasanjaya/card-authorization-service/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthorizationService interface {
	Authorize(req dto.AuthorizeRequest, idempotencyKey string) (*dto.AuthorizeApprovedResponse, *dto.AuthorizeDeclinedResponse, error)
	Reverse(authorizationID string) (*dto.ReverseResponse, error)
	GetTransactionHistory(cardID string) ([]dto.TransactionHistoryResponse, error)
}

type authorizationService struct {
	db       *gorm.DB
	cardRepo repository.CardRepository
	authRepo repository.AuthorizationRepository
}

func NewAuthorizationService(
	db *gorm.DB,
	cardRepo repository.CardRepository,
	authRepo repository.AuthorizationRepository,
) AuthorizationService {
	return &authorizationService{
		db:       db,
		cardRepo: cardRepo,
		authRepo: authRepo,
	}
}

// Authorize processes a card authorization request
func (s *authorizationService) Authorize(
	req dto.AuthorizeRequest,
	idempotencyKey string,
) (*dto.AuthorizeApprovedResponse, *dto.AuthorizeDeclinedResponse, error) {

	logger.Info("Authorization started",
		zap.String("cardNumber", maskCardNumber(req.CardNumber)),
		zap.String("merchantId", req.MerchantID),
		zap.Float64("amount", req.Amount),
		zap.String("currency", req.Currency),
	)

	// ── Idempotency Check ────────────────────────────────────────
	if idempotencyKey != "" {
		existing, err := s.authRepo.FindIdempotencyKey(idempotencyKey)
		if err == nil && existing != nil {
			// Key exists — return the same authorization
			auth, err := s.authRepo.FindByAuthorizationCode(existing.AuthorizationID.String())
			if err == nil && auth != nil {
				logger.Info("Idempotent response returned",
					zap.String("idempotencyKey", idempotencyKey),
				)
				if auth.Status == model.AuthorizationStatusApproved {
					return &dto.AuthorizeApprovedResponse{
						AuthorizationID:  auth.AuthorizationCode,
						Status:           string(auth.Status),
						RemainingBalance: 0, // already deducted
					}, nil, nil
				}
				return nil, &dto.AuthorizeDeclinedResponse{
					Status: string(auth.Status),
					Reason: "DUPLICATE_REQUEST",
				}, nil
			}
		}
	}

	// ── Step 1: Find Card ────────────────────────────────────────
	card, err := s.cardRepo.FindByCardNumber(req.CardNumber)
	if err != nil {
		logger.Info("Authorization declined - card not found",
			zap.String("cardNumber", maskCardNumber(req.CardNumber)),
		)
		return nil, &dto.AuthorizeDeclinedResponse{
			Status: "DECLINED",
			Reason: apperrors.CodeCardNotFound,
		}, nil
	}

	logger.Info("Card found", zap.String("cardId", card.ID.String()))

	// ── Step 2: Check Card Status ────────────────────────────────
	if card.Status == model.CardStatusFrozen {
		logger.Info("Authorization declined - card frozen",
			zap.String("cardId", card.ID.String()),
		)
		return nil, &dto.AuthorizeDeclinedResponse{
			Status: "DECLINED",
			Reason: apperrors.CodeCardFrozen,
		}, nil
	}

	// ── Step 3: Check Currency ───────────────────────────────────
	if !strings.EqualFold(card.Currency, req.Currency) {
		logger.Info("Authorization declined - currency mismatch",
			zap.String("cardCurrency", card.Currency),
			zap.String("requestCurrency", req.Currency),
		)
		return nil, &dto.AuthorizeDeclinedResponse{
			Status: "DECLINED",
			Reason: apperrors.CodeInvalidCurrency,
		}, nil
	}

	// ── Step 4-6: DB Transaction (lock, check balance, deduct) ───
	var approvedResponse *dto.AuthorizeApprovedResponse
	var declinedResponse *dto.AuthorizeDeclinedResponse

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		// Lock the card row to prevent race conditions
		lockedCard, err := s.cardRepo.FindByIDForUpdate(tx, card.ID)
		if err != nil {
			return apperrors.InternalError("failed to lock card")
		}

		// Check balance
		logger.Info("Checking balance",
			zap.Float64("balance", lockedCard.Balance),
			zap.Float64("amount", req.Amount),
		)

		if lockedCard.Balance < req.Amount {
			declinedResponse = &dto.AuthorizeDeclinedResponse{
				Status: "DECLINED",
				Reason: apperrors.CodeInsufficientFunds,
			}
			return nil // not a DB error, just a business decline
		}

		// Deduct balance
		lockedCard.Balance -= req.Amount
		if err := s.cardRepo.Update(lockedCard); err != nil {
			return apperrors.InternalError("failed to deduct balance")
		}

		// Save authorization
		authCode := generateAuthCode()
		auth := &model.Authorization{
			ID:                uuid.New(),
			AuthorizationCode: authCode,
			CardID:            lockedCard.ID,
			MerchantID:        req.MerchantID,
			MerchantName:      req.MerchantName,
			Amount:            req.Amount,
			Currency:          req.Currency,
			Status:            model.AuthorizationStatusApproved,
		}

		if err := s.authRepo.Create(tx, auth); err != nil {
			return apperrors.InternalError("failed to save authorization")
		}

		// Save idempotency key
		if idempotencyKey != "" {
			iKey := &model.IdempotencyKey{
				Key:             idempotencyKey,
				AuthorizationID: auth.ID,
			}
			if err := s.authRepo.SaveIdempotencyKey(tx, iKey); err != nil {
				return apperrors.InternalError("failed to save idempotency key")
			}
		}

		approvedResponse = &dto.AuthorizeApprovedResponse{
			AuthorizationID:  authCode,
			Status:           "APPROVED",
			RemainingBalance: lockedCard.Balance,
		}

		return nil
	})

	if txErr != nil {
		logger.Error("Authorization transaction failed", zap.Error(txErr))
		return nil, nil, txErr
	}

	if declinedResponse != nil {
		logger.Info("Authorization declined",
			zap.String("reason", declinedResponse.Reason),
		)
		return nil, declinedResponse, nil
	}

	logger.Info("Authorization approved",
		zap.String("authorizationId", approvedResponse.AuthorizationID),
		zap.Float64("remainingBalance", approvedResponse.RemainingBalance),
	)

	return approvedResponse, nil, nil
}

// Reverse restores the balance for an approved authorization
func (s *authorizationService) Reverse(authorizationID string) (*dto.ReverseResponse, error) {
	logger.Info("Reversal started",
		zap.String("authorizationId", authorizationID),
	)

	auth, err := s.authRepo.FindByAuthorizationCode(authorizationID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.AuthNotFound()
		}
		return nil, apperrors.InternalError("failed to find authorization")
	}

	if auth.Status == model.AuthorizationStatusReversed {
		return nil, apperrors.AuthAlreadyReversed()
	}

	var response *dto.ReverseResponse

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		// Lock the card row
		card, err := s.cardRepo.FindByIDForUpdate(tx, auth.CardID)
		if err != nil {
			return apperrors.InternalError("failed to lock card")
		}

		// Restore balance
		card.Balance += auth.Amount
		if err := s.cardRepo.Update(card); err != nil {
			return apperrors.InternalError("failed to restore balance")
		}

		// Update authorization status
		auth.Status = model.AuthorizationStatusReversed
		if err := s.authRepo.Update(tx, auth); err != nil {
			return apperrors.InternalError("failed to update authorization")
		}

		response = &dto.ReverseResponse{
			AuthorizationID: auth.AuthorizationCode,
			Status:          string(auth.Status),
			RefundedAmount:  auth.Amount,
		}

		return nil
	})

	if txErr != nil {
		logger.Error("Reversal transaction failed", zap.Error(txErr))
		return nil, txErr
	}

	logger.Info("Reversal successful",
		zap.String("authorizationId", authorizationID),
		zap.Float64("refundedAmount", response.RefundedAmount),
	)

	return response, nil
}

// GetTransactionHistory returns all transactions for a card
func (s *authorizationService) GetTransactionHistory(cardID string) ([]dto.TransactionHistoryResponse, error) {
	parsedID, err := uuid.Parse(cardID)
	if err != nil {
		return nil, apperrors.CardNotFound()
	}

	auths, err := s.authRepo.FindByCardID(parsedID)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch transactions")
	}

	var result []dto.TransactionHistoryResponse
	for _, auth := range auths {
		result = append(result, dto.TransactionHistoryResponse{
			AuthorizationID: auth.AuthorizationCode,
			MerchantName:    auth.MerchantName,
			Amount:          auth.Amount,
			Currency:        auth.Currency,
			Status:          string(auth.Status),
			CreatedAt:       auth.CreatedAt.Format(time.RFC3339),
		})
	}

	return result, nil
}

// ── Helpers ──────────────────────────────────────────────────────

// generateAuthCode creates a unique authorization code e.g. AUTH-A1B2C3D4
func generateAuthCode() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 8)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return fmt.Sprintf("AUTH-%s", string(code))
}

// maskCardNumber hides middle digits for safe logging e.g. 4556****1234
func maskCardNumber(cardNumber string) string {
	if len(cardNumber) < 8 {
		return "****"
	}
	return cardNumber[:4] + "****" + cardNumber[len(cardNumber)-4:]
}
