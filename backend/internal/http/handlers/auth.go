package handlers

import (
	"net/http"

	"mini-banking-platform/internal/http/dto"
	"mini-banking-platform/internal/http/response"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	handler *Handler
}

func NewAuthHandler(h *Handler) *AuthHandler {
	return &AuthHandler{handler: h}
}

func (h *AuthHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WithBindError(c, err)
		return
	}

	resp, err := h.handler.authService.Register(ctx, req, h.handler.config.InitialBalanceUSDCents, h.handler.config.InitialBalanceEURCents)
	if err != nil {
		response.WithServiceError(c, err)
		return
	}

	response.WithJSON(c, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WithBindError(c, err)
		return
	}

	resp, err := h.handler.authService.Login(ctx, req)
	if err != nil {
		response.WithServiceError(c, err)
		return
	}

	response.WithJSON(c, http.StatusOK, resp)
}

func (h *AuthHandler) GetMe(c *gin.Context) {
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
	user, err := h.handler.authService.GetUser(ctx, userIDStr)
	if err != nil {
		response.WithServiceError(c, err)
		return
	}

	response.WithJSON(c, http.StatusOK, user)
}

