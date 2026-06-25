package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/storedge/storedge/services/search-match/internal/service"
)

type InventoryHandler struct {
	svc    *service.DashboardService
	logger *zap.Logger
}

func NewInventoryHandler(svc *service.DashboardService, logger *zap.Logger) *InventoryHandler {
	return &InventoryHandler{svc: svc, logger: logger}
}

// ListPallets handles GET /api/v1/inventory/pallets
func (h *InventoryHandler) ListPallets(c *gin.Context) {
	tenantID := currentUserID(c)
	pallets, err := h.svc.ListPallets(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("list pallets failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load inventory"})
		return
	}
	if pallets == nil {
		pallets = []service.PalletItem{}
	}
	c.JSON(http.StatusOK, gin.H{"items": pallets})
}

// InitiateRelease handles POST /api/v1/inventory/pallets/:id/release/initiate
func (h *InventoryHandler) InitiateRelease(c *gin.Context) {
	palletID := c.Param("id")
	requesterID := currentUserID(c)
	if err := h.svc.InitiateRelease(c.Request.Context(), palletID, requesterID); err != nil {
		h.logger.Error("initiate release failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OTP sent to registered WhatsApp number"})
}

// AuthorizeRelease handles POST /api/v1/inventory/pallets/:id/release/authorize
func (h *InventoryHandler) AuthorizeRelease(c *gin.Context) {
	palletID := c.Param("id")
	var body struct {
		OTP string `json:"otp" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "otp must be 6 digits"})
		return
	}
	if err := h.svc.AuthorizeRelease(c.Request.Context(), palletID, body.OTP); err != nil {
		h.logger.Error("authorize release failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Release authorized. Operator will prepare goods for pickup."})
}
