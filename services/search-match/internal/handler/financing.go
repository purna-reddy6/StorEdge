package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/storedge/storedge/services/search-match/internal/service"
)

type FinancingHandler struct {
	svc    *service.DashboardService
	logger *zap.Logger
}

func NewFinancingHandler(svc *service.DashboardService, logger *zap.Logger) *FinancingHandler {
	return &FinancingHandler{svc: svc, logger: logger}
}

// ListReceipts handles GET /api/v1/financing/receipts
func (h *FinancingHandler) ListReceipts(c *gin.Context) {
	tenantID := tenantFilterID(c)
	receipts, err := h.svc.ListReceipts(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("list receipts failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load receipts"})
		return
	}
	if receipts == nil {
		receipts = []service.EnwrReceipt{}
	}
	c.JSON(http.StatusOK, gin.H{"receipts": receipts})
}

// ApplyForLoan handles POST /api/v1/financing/receipts/:id/loans
func (h *FinancingHandler) ApplyForLoan(c *gin.Context) {
	receiptID := c.Param("id")
	applicantID := currentUserID(c)
	if err := h.svc.ApplyForLoan(c.Request.Context(), receiptID, applicantID); err != nil {
		h.logger.Error("loan application failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Loan application submitted. Partner bank will contact you within 24 hours.",
		"psl_limit_inr": 7500000,
		"origination_fee_pct": 1.5,
	})
}
