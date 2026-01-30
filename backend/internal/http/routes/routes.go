package routes

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"mini-banking-platform/internal/http/handlers"
	"mini-banking-platform/internal/http/middleware"
	"mini-banking-platform/internal/jwt"

	"github.com/gin-gonic/gin"
)

func NewRouter(handler *handlers.Handler, jwtService *jwt.Service, logger *slog.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	allowOrigin := os.Getenv("CORS_ALLOW_ORIGIN")
	if allowOrigin == "" {
		allowOrigin = "http://localhost:3000"
	}
	allowOrigins := strings.Split(allowOrigin, ",")

	router.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowedOrigin := "*"
		if allowOrigin != "*" && origin != "" {
			for _, o := range allowOrigins {
				if strings.TrimSpace(o) == origin {
					allowedOrigin = origin
					break
				}
			}
		} else if allowOrigin == "*" {
			if origin != "" {
				allowedOrigin = origin
			} else {
				allowedOrigin = "*"
			}
		}
		

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		
		if allowedOrigin != "*" {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Vary", "Origin")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	authHandler := handlers.NewAuthHandler(handler)
	accountHandler := handlers.NewAccountHandler(handler)
	transactionHandler := handlers.NewTransactionHandler(handler)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthMiddleware(jwtService), authHandler.GetMe)
		}

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtService))
		{
			protected.GET("/accounts", accountHandler.GetAccounts)
			protected.GET("/accounts/:id/balance", accountHandler.GetBalance)
			protected.GET("/accounts/reconcile", accountHandler.ReconcileBalances)

			protected.POST("/transactions/transfer", transactionHandler.Transfer)
			protected.POST("/transactions/exchange", transactionHandler.Exchange)
			protected.GET("/transactions", transactionHandler.GetTransactions)
		}
	}

	return router
}
