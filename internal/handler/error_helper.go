package handler

import (
	"net/http"

	apperrors "github.com/dewadityasanjaya/card-authorization-service/internal/errors"
	"github.com/gin-gonic/gin"
)

// handleError maps AppError codes to HTTP status codes
func handleError(c *gin.Context, err error) {
	appErr, ok := apperrors.IsAppError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": apperrors.CodeInternalError,
		})
		return
	}

	status := toHTTPStatus(appErr.Code)
	c.JSON(status, gin.H{
		"error":   appErr.Code,
		"message": appErr.Message,
	})
}

func toHTTPStatus(code string) int {
	switch code {
	case apperrors.CodeCardNotFound, apperrors.CodeAuthNotFound:
		return http.StatusNotFound
	case apperrors.CodeCardFrozen,
		apperrors.CodeInsufficientFunds,
		apperrors.CodeInvalidCurrency,
		apperrors.CodeInvalidAmount,
		apperrors.CodeAuthAlreadyReversed,
		apperrors.CodeCannotFreeze,
		apperrors.CodeCannotUnfreeze:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
