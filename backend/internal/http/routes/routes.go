package routes

import (
	"log/slog"

	"mini-banking-platform/internal/http/handlers"
	"mini-banking-platform/internal/http/middleware"
	"mini-banking-platform/internal/jwt"
	"github.com/gin-gonic/gin"
)

func NewRouter(handler *handlers.Handler, jwtService *jwt.Service, logger *slog.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
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

