package handlers

import (
	"net/http"

	"mini-banking-platform/internal/http/dto"
	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	handler *Handler
}

func NewTransactionHandler(h *Handler) *TransactionHandler {
	return &TransactionHandler{handler: h}
}

func (h *TransactionHandler) Transfer(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondWithError(c, "user not authenticated", http.StatusUnauthorized)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		respondWithError(c, "invalid user ID", http.StatusInternalServerError)
		return
	}

	var req dto.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithBindError(c, err)
		return
	}

	ctx := c.Request.Context()
	transaction, err := h.handler.transactionService.Transfer(ctx, userIDStr, req)
	if err != nil {
		respondWithServiceError(c, err)
		return
	}

	respondWithJSON(c, http.StatusCreated, transaction)
}

func (h *TransactionHandler) Exchange(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondWithError(c, "user not authenticated", http.StatusUnauthorized)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		respondWithError(c, "invalid user ID", http.StatusInternalServerError)
		return
	}

	var req dto.ExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithBindError(c, err)
		return
	}

	ctx := c.Request.Context()
	transaction, err := h.handler.transactionService.Exchange(ctx, userIDStr, req)
	if err != nil {
		respondWithServiceError(c, err)
		return
	}

	respondWithJSON(c, http.StatusCreated, transaction)
}

func (h *TransactionHandler) GetTransactions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondWithError(c, "user not authenticated", http.StatusUnauthorized)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		respondWithError(c, "invalid user ID", http.StatusInternalServerError)
		return
	}

	var req dto.GetTransactionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondWithBindError(c, err)
		return
	}

	page := req.Page
	if page < 1 {
		page = h.handler.config.DefaultPage
	}

	limit := req.Limit
	if limit < 1 {
		limit = h.handler.config.DefaultLimit
	}
	if limit > h.handler.config.MaxLimit {
		limit = h.handler.config.MaxLimit
	}

	ctx := c.Request.Context()
	transactions, total, err := h.handler.transactionService.GetTransactions(ctx, userIDStr, req.Type, page, limit)
	if err != nil {
		respondWithServiceError(c, err)
		return
	}

	respondWithJSON(c, http.StatusOK, gin.H{
		"transactions": transactions,
		"page":         page,
		"limit":        limit,
		"total":        total,
	})
}

