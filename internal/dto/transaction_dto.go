package dto

// ── Request DTOs ─────────────────────────────────

type AuthorizeRequest struct {
	CardNumber   string  `json:"cardNumber" binding:"required"`
	MerchantID   string  `json:"merchantId" binding:"required"`
	MerchantName string  `json:"merchantName" binding:"required"`
	Currency     string  `json:"currency" binding:"required,len=3"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
}

// ── Response DTOs ────────────────────────────────

type AuthorizeApprovedResponse struct {
	AuthorizationID  string  `json:"authorizationId"`
	Status           string  `json:"status"`
	RemainingBalance float64 `json:"remainingBalance"`
}

type AuthorizeDeclinedResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type ReverseResponse struct {
	AuthorizationID string  `json:"authorizationId"`
	Status          string  `json:"status"`
	RefundedAmount  float64 `json:"refundedAmount"`
}

type TransactionHistoryResponse struct {
	AuthorizationID string  `json:"authorizationId"`
	MerchantName    string  `json:"merchant"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Status          string  `json:"status"`
	CreatedAt       string  `json:"createdAt"`
}
