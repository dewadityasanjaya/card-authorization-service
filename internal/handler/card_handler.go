package handler

import (
	"net/http"

	"github.com/dewadityasanjaya/card-authorization-service/internal/dto"
	"github.com/dewadityasanjaya/card-authorization-service/internal/service"
	"github.com/gin-gonic/gin"
)

type CardHandler struct {
	cardService service.CardService
}

func NewCardHandler(cardService service.CardService) *CardHandler {
	return &CardHandler{cardService: cardService}
}

// CreateCard godoc
// POST /cards
func (h *CardHandler) CreateCard(c *gin.Context) {
	var req dto.CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := h.cardService.CreateCard(req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetCard godoc
// GET /cards/:id
func (h *CardHandler) GetCard(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.cardService.GetCard(id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// FreezeCard godoc
// POST /cards/:id/freeze
func (h *CardHandler) FreezeCard(c *gin.Context) {
	id := c.Param("id")

	if err := h.cardService.FreezeCard(id); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "card frozen successfully",
	})
}

// UnfreezeCard godoc
// POST /cards/:id/unfreeze
func (h *CardHandler) UnfreezeCard(c *gin.Context) {
	id := c.Param("id")

	if err := h.cardService.UnfreezeCard(id); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "card unfrozen successfully",
	})
}

// TopUp godoc
// POST /cards/:id/topup
func (h *CardHandler) TopUp(c *gin.Context) {
	id := c.Param("id")

	var req dto.TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := h.cardService.TopUp(id, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
