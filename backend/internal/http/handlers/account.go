package handlers

import (
	"net/http"

	"mini-banking-platform/internal/http/response"
	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	handler *Handler
}

func NewAccountHandler(h *Handler) *AccountHandler {
	return &AccountHandler{handler: h}
}

func (h *AccountHandler) GetAccounts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.WithError(c, "user not authenticated", http.StatusUnauthorized)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		response.WithError(c, "invalid user ID", http.StatusInternalServerError)
		return
	}

	ctx := c.Request.Context()
	accounts, err := h.handler.accountService.GetUserAccounts(ctx, userIDStr)
	if err != nil {
		response.WithServiceError(c, err)
		return
	}

	response.WithJSON(c, http.StatusOK, gin.H{"accounts": accounts})
}

func (h *AccountHandler) GetBalance(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.WithError(c, "user not authenticated", http.StatusUnauthorized)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		response.WithError(c, "invalid user ID", http.StatusInternalServerError)
		return
	}

	accountID := c.Param("id")

	ctx := c.Request.Context()
	account, err := h.handler.accountService.GetAccountBalance(ctx, userIDStr, accountID)
	if err != nil {
		response.WithServiceError(c, err)
		return
	}

	response.WithJSON(c, http.StatusOK, gin.H{
		"balance_cents": account.BalanceCents,
		"currency":      account.Currency,
	})
}

func (h *AccountHandler) ReconcileBalances(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.WithError(c, "user not authenticated", http.StatusUnauthorized)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		response.WithError(c, "invalid user ID", http.StatusInternalServerError)
		return
	}

	ctx := c.Request.Context()
	results, err := h.handler.accountService.ReconcileBalances(ctx, userIDStr)
	if err != nil {
		response.WithServiceError(c, err)
		return
	}

	response.WithJSON(c, http.StatusOK, results)
}

