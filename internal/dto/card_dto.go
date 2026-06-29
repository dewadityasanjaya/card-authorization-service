package dto

// ── Request DTOs ─────────────────────────────────

type CreateCardRequest struct {
	CardholderName string  `json:"cardholderName" binding:"required"`
	Currency       string  `json:"currency" binding:"required,len=3"`
	InitialBalance float64 `json:"initialBalance" binding:"min=0"`
}

type TopUpRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

// ── Response DTOs ────────────────────────────────

type CreateCardResponse struct {
	CardID     string  `json:"cardId"`
	CardNumber string  `json:"cardNumber"`
	Status     string  `json:"status"`
	Balance    float64 `json:"balance"`
}

type GetCardResponse struct {
	ID             string  `json:"id"`
	CardholderName string  `json:"cardholderName"`
	CardNumber     string  `json:"cardNumber"`
	Status         string  `json:"status"`
	Balance        float64 `json:"balance"`
	Currency       string  `json:"currency"`
}

type TopUpResponse struct {
	CardID   string  `json:"cardId"`
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}
