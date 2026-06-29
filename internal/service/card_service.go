package service

import (
	"github.com/dewadityasanjaya/card-authorization-service/internal/dto"
	apperrors "github.com/dewadityasanjaya/card-authorization-service/internal/errors"
	"github.com/dewadityasanjaya/card-authorization-service/internal/model"
	"github.com/dewadityasanjaya/card-authorization-service/internal/repository"
	"github.com/dewadityasanjaya/card-authorization-service/pkg/cardnumber"
	"github.com/dewadityasanjaya/card-authorization-service/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CardService interface {
	CreateCard(req dto.CreateCardRequest) (*dto.CreateCardResponse, error)
	GetCard(id string) (*dto.GetCardResponse, error)
	FreezeCard(id string) error
	UnfreezeCard(id string) error
	TopUp(id string, req dto.TopUpRequest) (*dto.TopUpResponse, error)
}

type cardService struct {
	cardRepo repository.CardRepository
}

func NewCardService(cardRepo repository.CardRepository) CardService {
	return &cardService{cardRepo: cardRepo}
}

// CreateCard creates a new prepaid card
func (s *cardService) CreateCard(req dto.CreateCardRequest) (*dto.CreateCardResponse, error) {
	logger.Info("Creating new card",
		zap.String("cardholderName", req.CardholderName),
		zap.String("currency", req.Currency),
	)

	card := &model.Card{
		ID:             uuid.New(),
		CardNumber:     cardnumber.Generate(),
		CardholderName: req.CardholderName,
		Status:         model.CardStatusActive,
		Currency:       req.Currency,
		Balance:        req.InitialBalance,
	}

	if err := s.cardRepo.Create(card); err != nil {
		logger.Error("Failed to create card", zap.Error(err))
		return nil, apperrors.InternalError("failed to create card")
	}

	logger.Info("Card created successfully",
		zap.String("cardId", card.ID.String()),
		zap.String("cardNumber", card.CardNumber),
	)

	return &dto.CreateCardResponse{
		CardID:     card.ID.String(),
		CardNumber: card.CardNumber,
		Status:     string(card.Status),
		Balance:    card.Balance,
	}, nil
}

// GetCard retrieves a card by ID
func (s *cardService) GetCard(id string) (*dto.GetCardResponse, error) {
	cardID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.CardNotFound()
	}

	card, err := s.cardRepo.FindByID(cardID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.CardNotFound()
		}
		logger.Error("Failed to find card", zap.Error(err))
		return nil, apperrors.InternalError("failed to find card")
	}

	return &dto.GetCardResponse{
		ID:             card.ID.String(),
		CardholderName: card.CardholderName,
		CardNumber:     card.CardNumber,
		Status:         string(card.Status),
		Balance:        card.Balance,
		Currency:       card.Currency,
	}, nil
}

// FreezeCard changes card status from ACTIVE to FROZEN
func (s *cardService) FreezeCard(id string) error {
	cardID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.CardNotFound()
	}

	card, err := s.cardRepo.FindByID(cardID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.CardNotFound()
		}
		return apperrors.InternalError("failed to find card")
	}

	if card.Status == model.CardStatusFrozen {
		return apperrors.CannotFreeze()
	}

	card.Status = model.CardStatusFrozen

	if err := s.cardRepo.Update(card); err != nil {
		logger.Error("Failed to freeze card", zap.Error(err))
		return apperrors.InternalError("failed to freeze card")
	}

	logger.Info("Card frozen", zap.String("cardId", id))
	return nil
}

// UnfreezeCard changes card status from FROZEN to ACTIVE
func (s *cardService) UnfreezeCard(id string) error {
	cardID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.CardNotFound()
	}

	card, err := s.cardRepo.FindByID(cardID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.CardNotFound()
		}
		return apperrors.InternalError("failed to find card")
	}

	if card.Status == model.CardStatusActive {
		return apperrors.CannotUnfreeze()
	}

	card.Status = model.CardStatusActive

	if err := s.cardRepo.Update(card); err != nil {
		logger.Error("Failed to unfreeze card", zap.Error(err))
		return apperrors.InternalError("failed to unfreeze card")
	}

	logger.Info("Card unfrozen", zap.String("cardId", id))
	return nil
}

// TopUp adds balance to an active card
func (s *cardService) TopUp(id string, req dto.TopUpRequest) (*dto.TopUpResponse, error) {
	cardID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.CardNotFound()
	}

	card, err := s.cardRepo.FindByID(cardID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.CardNotFound()
		}
		return nil, apperrors.InternalError("failed to find card")
	}

	if card.Status == model.CardStatusFrozen {
		return nil, apperrors.CardFrozen()
	}

	card.Balance += req.Amount

	if err := s.cardRepo.Update(card); err != nil {
		logger.Error("Failed to top up card", zap.Error(err))
		return nil, apperrors.InternalError("failed to top up card")
	}

	logger.Info("Card topped up",
		zap.String("cardId", id),
		zap.Float64("amount", req.Amount),
		zap.Float64("newBalance", card.Balance),
	)

	return &dto.TopUpResponse{
		CardID:   card.ID.String(),
		Balance:  card.Balance,
		Currency: card.Currency,
	}, nil
}
