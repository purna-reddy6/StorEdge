package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/storedge/storedge/services/search-match/internal/service"
)

type DashboardHandler struct {
	svc    *service.DashboardService
	logger *zap.Logger
}

func NewDashboardHandler(svc *service.DashboardService, logger *zap.Logger) *DashboardHandler {
	return &DashboardHandler{svc: svc, logger: logger}
}

// GetOccupancy handles GET /api/v1/operator/occupancy
func (h *DashboardHandler) GetOccupancy(c *gin.Context) {
	ownerID := currentUserID(c)
	stats, err := h.svc.GetOccupancyStats(c.Request.Context(), ownerID)
	if err != nil {
		h.logger.Error("occupancy stats failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load occupancy"})
		return
	}
	if stats == nil {
		stats = []service.OccupancyStat{}
	}
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// GetOperatorBookings handles GET /api/v1/operator/bookings
func (h *DashboardHandler) GetOperatorBookings(c *gin.Context) {
	ownerID := currentUserID(c)
	bookings, err := h.svc.GetOccupancyStats(c.Request.Context(), ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load bookings"})
		return
	}
	// Reuse occupancy query for now; operator booking list is via the standard /bookings endpoint
	_ = bookings
	c.JSON(http.StatusOK, gin.H{"bookings": []any{}})
}

// GetAlerts handles GET /api/v1/operator/alerts
func (h *DashboardHandler) GetAlerts(c *gin.Context) {
	ownerID := currentUserID(c)
	alerts, err := h.svc.ListAlerts(c.Request.Context(), ownerID)
	if err != nil {
		h.logger.Error("list alerts failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load alerts"})
		return
	}
	if alerts == nil {
		alerts = []service.IoTAlert{}
	}
	c.JSON(http.StatusOK, gin.H{"alerts": alerts})
}
