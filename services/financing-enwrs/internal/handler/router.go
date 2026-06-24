package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/storedge/storedge/services/financing-enwrs/internal/service"
)

func NewRouter(enwrsSvc *service.ENWRsService, logger *zap.Logger) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "financing-enwrs"})
	})

	h := NewENWRsHandler(enwrsSvc, logger)
	api := r.Group("/api/v1")
	{
		api.POST("/enwrs/receipts", h.CreateReceipt)
		api.GET("/enwrs/receipts/:id", h.GetReceipt)
		api.POST("/enwrs/receipts/:id/issue", h.IssueReceipt)
		api.GET("/enwrs/depositor/:depositorId/receipts", h.ListReceiptsByDepositor)

		api.POST("/enwrs/loans", h.ApplyForLoan)
		api.GET("/enwrs/loans/:id", h.GetLoan)
	}

	return r
}

type ENWRsHandler struct {
	svc    *service.ENWRsService
	logger *zap.Logger
}

func NewENWRsHandler(svc *service.ENWRsService, logger *zap.Logger) *ENWRsHandler {
	return &ENWRsHandler{svc: svc, logger: logger}
}

func (h *ENWRsHandler) CreateReceipt(c *gin.Context) {
	var req service.CreateReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	receipt, err := h.svc.CreateReceipt(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("create receipt failed", zap.Error(err))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"receipt": receipt})
}

func (h *ENWRsHandler) GetReceipt(c *gin.Context) {
	receipt, err := h.svc.GetReceipt(c.Request.Context(), c.Param("id"))
	if err != nil || receipt == nil {
		c.JSON(404, gin.H{"error": "receipt not found"})
		return
	}
	c.JSON(200, gin.H{"receipt": receipt})
}

func (h *ENWRsHandler) IssueReceipt(c *gin.Context) {
	receipt, err := h.svc.IssueReceipt(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"receipt": receipt, "message": "e-NWR issued and registered with NERL"})
}

func (h *ENWRsHandler) ListReceiptsByDepositor(c *gin.Context) {
	receipts, err := h.svc.ListReceiptsByDepositor(c.Request.Context(), c.Param("depositorId"))
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch receipts"})
		return
	}
	c.JSON(200, gin.H{"receipts": receipts, "count": len(receipts)})
}

func (h *ENWRsHandler) ApplyForLoan(c *gin.Context) {
	var req service.LoanApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	loan, err := h.svc.ApplyForLoan(c.Request.Context(), req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{
		"loan":    loan,
		"message": "Loan application submitted. Under bank review.",
	})
}

func (h *ENWRsHandler) GetLoan(c *gin.Context) {
	loan, err := h.svc.GetLoan(c.Request.Context(), c.Param("id"))
	if err != nil || loan == nil {
		c.JSON(404, gin.H{"error": "loan not found"})
		return
	}
	c.JSON(200, gin.H{"loan": loan})
}
