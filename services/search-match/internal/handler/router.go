package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/storedge/storedge/services/search-match/internal/service"
)

func NewRouter(
	matchingSvc *service.MatchingService,
	bookingSvc *service.BookingService,
	authSvc *service.AuthService,
	jwtSecret string,
	logger *zap.Logger,
) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(requestLogger(logger))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "search-match"})
	})

	api := r.Group("/api/v1")

	// Auth routes (no JWT required)
	auth := api.Group("/auth")
	{
		authHandler := NewAuthHandler(authSvc, logger)
		auth.POST("/otp/request", authHandler.RequestOTP)
		auth.POST("/otp/verify", authHandler.VerifyOTP)
	}

	// Protected routes
	protected := api.Group("/")
	protected.Use(jwtMiddleware(authSvc))
	{
		wh := NewWarehouseHandler(matchingSvc, logger)
		protected.GET("/warehouses/search", wh.Search)
		protected.GET("/warehouses/:id", wh.GetWarehouse)
		protected.GET("/warehouses/:id/price", wh.GetDynamicPrice)

		bk := NewBookingHandler(bookingSvc, logger)
		protected.POST("/bookings", bk.CreateBooking)
		protected.GET("/bookings/:id", bk.GetBooking)
	}

	return r
}

func jwtMiddleware(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		user, err := authSvc.ValidateJWT(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func requestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logger.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
		)
	}
}
