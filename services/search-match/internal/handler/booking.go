package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/storedge/storedge/services/search-match/internal/matching"
	"github.com/storedge/storedge/services/search-match/internal/service"
)

type BookingHandler struct {
	bookingSvc *service.BookingService
	logger     *zap.Logger
}

func NewBookingHandler(bookingSvc *service.BookingService, logger *zap.Logger) *BookingHandler {
	return &BookingHandler{bookingSvc: bookingSvc, logger: logger}
}

// CreateBooking handles POST /api/v1/bookings
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req matching.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Inject authenticated user's ID if not set
	if req.TenantID == "" {
		user, _ := c.Get("user")
		if u, ok := user.(*service.User); ok {
			req.TenantID = u.ID
		}
	}

	booking, err := h.bookingSvc.CreateBooking(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("create booking failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"booking": booking,
		"message": "Booking created. Complete payment to confirm.",
	})
}

// ListBookings handles GET /api/v1/bookings
func (h *BookingHandler) ListBookings(c *gin.Context) {
	userID := currentUserID(c)
	bookings, err := h.bookingSvc.ListBookings(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("list bookings failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bookings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

// GetBooking handles GET /api/v1/bookings/:id
func (h *BookingHandler) GetBooking(c *gin.Context) {
	id := c.Param("id")
	booking, err := h.bookingSvc.GetBooking(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch booking"})
		return
	}
	if booking == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking": booking})
}
