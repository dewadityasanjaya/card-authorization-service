package handler

import (
	"net/http"

	"github.com/dewadityasanjaya/card-authorization-service/internal/dto"
	"github.com/dewadityasanjaya/card-authorization-service/internal/service"
	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	authService service.AuthorizationService
}

func NewTransactionHandler(authService service.AuthorizationService) *TransactionHandler {
	return &TransactionHandler{authService: authService}
}

// Authorize godoc
// POST /transactions/authorize
func (h *TransactionHandler) Authorize(c *gin.Context) {
	var req dto.AuthorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get idempotency key from header
	idempotencyKey := c.GetHeader("Idempotency-Key")

	approved, declined, err := h.authService.Authorize(req, idempotencyKey)
	if err != nil {
		handleError(c, err)
		return
	}

	if declined != nil {
		c.JSON(http.StatusOK, declined)
		return
	}

	c.JSON(http.StatusOK, approved)
}

// Reverse godoc
// POST /transactions/:authorizationId/reverse
func (h *TransactionHandler) Reverse(c *gin.Context) {
	authorizationID := c.Param("authorizationId")

	resp, err := h.authService.Reverse(authorizationID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetTransactionHistory godoc
// GET /cards/:id/transactions
func (h *TransactionHandler) GetTransactionHistory(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.authService.(service.AuthorizationService).GetTransactionHistory(id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
