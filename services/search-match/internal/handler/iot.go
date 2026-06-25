package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/storedge/storedge/services/search-match/internal/service"
)

type IoTHandler struct {
	svc    *service.DashboardService
	logger *zap.Logger
}

func NewIoTHandler(svc *service.DashboardService, logger *zap.Logger) *IoTHandler {
	return &IoTHandler{svc: svc, logger: logger}
}

// ListAlerts handles GET /api/v1/iot/alerts
func (h *IoTHandler) ListAlerts(c *gin.Context) {
	ownerID := tenantFilterID(c)
	alerts, err := h.svc.ListAlerts(c.Request.Context(), ownerID)
	if err != nil {
		h.logger.Error("list IoT alerts failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load alerts"})
		return
	}
	if alerts == nil {
		alerts = []service.IoTAlert{}
	}
	c.JSON(http.StatusOK, gin.H{"alerts": alerts})
}

// ResolveAlert handles PATCH /api/v1/iot/alerts/:id/resolve
func (h *IoTHandler) ResolveAlert(c *gin.Context) {
	alertID := c.Param("id")
	if err := h.svc.ResolveAlert(c.Request.Context(), alertID); err != nil {
		h.logger.Error("resolve alert failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve alert"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "alert resolved"})
}
