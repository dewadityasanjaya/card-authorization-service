package errors

import "errors"

// App error codes
const (
	CodeCardNotFound        = "CARD_NOT_FOUND"
	CodeCardFrozen          = "CARD_FROZEN"
	CodeInsufficientFunds   = "INSUFFICIENT_FUNDS"
	CodeInvalidCurrency     = "INVALID_CURRENCY"
	CodeInvalidAmount       = "INVALID_AMOUNT"
	CodeAuthAlreadyReversed = "AUTH_ALREADY_REVERSED"
	CodeAuthNotFound        = "AUTH_NOT_FOUND"
	CodeCannotFreeze        = "CANNOT_FREEZE"
	CodeCannotUnfreeze      = "CANNOT_UNFREEZE"
	CodeInternalError       = "INTERNAL_ERROR"
)

// AppError is a custom error that carries a code and message
type AppError struct {
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

// Constructors
func NewAppError(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func CardNotFound() *AppError {
	return NewAppError(CodeCardNotFound, "card not found")
}

func CardFrozen() *AppError {
	return NewAppError(CodeCardFrozen, "card is frozen")
}

func InsufficientFunds() *AppError {
	return NewAppError(CodeInsufficientFunds, "insufficient funds")
}

func InvalidCurrency() *AppError {
	return NewAppError(CodeInvalidCurrency, "currency mismatch")
}

func InvalidAmount() *AppError {
	return NewAppError(CodeInvalidAmount, "amount must be greater than 0")
}

func AuthAlreadyReversed() *AppError {
	return NewAppError(CodeAuthAlreadyReversed, "authorization already reversed")
}

func AuthNotFound() *AppError {
	return NewAppError(CodeAuthNotFound, "authorization not found")
}

func CannotFreeze() *AppError {
	return NewAppError(CodeCannotFreeze, "card is already frozen")
}

func CannotUnfreeze() *AppError {
	return NewAppError(CodeCannotUnfreeze, "card is already active")
}

func InternalError(msg string) *AppError {
	return NewAppError(CodeInternalError, msg)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
