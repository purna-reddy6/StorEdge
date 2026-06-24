package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/storedge/storedge/services/search-match/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
	logger  *zap.Logger
}

func NewAuthHandler(authSvc *service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, logger: logger}
}

type otpRequestBody struct {
	Phone   string `json:"phone"   binding:"required"`
	Purpose string `json:"purpose"`
}

type otpVerifyBody struct {
	Phone string `json:"phone" binding:"required"`
	OTP   string `json:"otp"   binding:"required"`
}

// RequestOTP handles POST /api/v1/auth/otp/request
func (h *AuthHandler) RequestOTP(c *gin.Context) {
	var body otpRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	purpose := body.Purpose
	if purpose == "" {
		purpose = "login"
	}

	otp, err := h.authSvc.RequestOTP(c.Request.Context(), body.Phone, purpose)
	if err != nil {
		h.logger.Error("OTP request failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate OTP"})
		return
	}

	resp := gin.H{"message": "OTP sent via SMS", "expires_in": 600}

	// In development, include OTP in response for testing
	resp["dev_otp"] = otp

	c.JSON(http.StatusOK, resp)
}

// VerifyOTP handles POST /api/v1/auth/otp/verify
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var body otpVerifyBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.authSvc.VerifyOTPAndLogin(c.Request.Context(), body.Phone, body.OTP)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}
