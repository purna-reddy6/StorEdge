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
	dashboardSvc *service.DashboardService,
	jwtSecret string,
	logger *zap.Logger,
) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(requestLogger(logger))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "search-match"})
	})

	api := r.Group("/api/v1")

	// Auth — support both /otp/send and /otp/request for client compatibility
	auth := api.Group("/auth")
	{
		authHandler := NewAuthHandler(authSvc, logger)
		auth.POST("/otp/send", authHandler.RequestOTP)
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
		protected.GET("/bookings", bk.ListBookings)
		protected.GET("/bookings/:id", bk.GetBooking)

		db := NewDashboardHandler(dashboardSvc, logger)
		protected.GET("/operator/occupancy", db.GetOccupancy)
		protected.GET("/operator/bookings", db.GetOperatorBookings)
		protected.GET("/operator/alerts", db.GetAlerts)

		iv := NewInventoryHandler(dashboardSvc, logger)
		protected.GET("/inventory/pallets", iv.ListPallets)
		protected.POST("/inventory/pallets/:id/release/initiate", iv.InitiateRelease)
		protected.POST("/inventory/pallets/:id/release/authorize", iv.AuthorizeRelease)

		fn := NewFinancingHandler(dashboardSvc, logger)
		protected.GET("/financing/receipts", fn.ListReceipts)
		protected.POST("/financing/receipts/:id/loans", fn.ApplyForLoan)

		iot := NewIoTHandler(dashboardSvc, logger)
		protected.GET("/iot/alerts", iot.ListAlerts)
		protected.PATCH("/iot/alerts/:id/resolve", iot.ResolveAlert)
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
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
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

func currentUserID(c *gin.Context) string {
	if u, ok := c.Get("user"); ok {
		if user, ok := u.(*service.User); ok {
			return user.ID
		}
	}
	return ""
}
