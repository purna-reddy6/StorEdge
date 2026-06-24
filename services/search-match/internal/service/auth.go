package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type User struct {
	ID       string `json:"id"`
	Phone    string `json:"phone"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Language string `json:"language_pref"`
}

type AuthService struct {
	db        *sql.DB
	jwtSecret []byte
	logger    *zap.Logger
}

func NewAuthService(db *sql.DB, jwtSecret string, logger *zap.Logger) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: []byte(jwtSecret),
		logger:    logger,
	}
}

// RequestOTP generates a 6-digit OTP for a phone number and stores the hash.
// In dev mode, OTP is returned in response (not sent via SMS).
func (s *AuthService) RequestOTP(ctx context.Context, phone, purpose string) (string, error) {
	otp, err := generateOTP(6)
	if err != nil {
		return "", fmt.Errorf("generate otp: %w", err)
	}

	otpHash := hashOTP(otp)
	expiresAt := time.Now().Add(10 * time.Minute)

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO otp_requests (phone, otp_hash, purpose, expires_at)
		VALUES ($1, $2, $3, $4)`,
		phone, otpHash, purpose, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("store otp: %w", err)
	}

	s.logger.Info("OTP generated", zap.String("phone", phone), zap.String("purpose", purpose))
	// Production: send via SMS gateway. Dev: return in response.
	return otp, nil
}

// VerifyOTPAndLogin validates the OTP, creates/fetches user, returns JWT.
func (s *AuthService) VerifyOTPAndLogin(ctx context.Context, phone, otp string) (string, *User, error) {
	otpHash := hashOTP(otp)

	var requestID string
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM otp_requests
		WHERE phone = $1 AND otp_hash = $2 AND purpose = 'login'
		  AND expires_at > NOW() AND used_at IS NULL
		ORDER BY created_at DESC LIMIT 1`,
		phone, otpHash,
	).Scan(&requestID)
	if err == sql.ErrNoRows {
		return "", nil, fmt.Errorf("invalid or expired OTP")
	}
	if err != nil {
		return "", nil, fmt.Errorf("verify otp: %w", err)
	}

	// Mark OTP as used
	_, err = s.db.ExecContext(ctx,
		"UPDATE otp_requests SET used_at = NOW() WHERE id = $1", requestID)
	if err != nil {
		s.logger.Warn("failed to mark OTP used", zap.Error(err))
	}

	// Upsert user
	user := &User{}
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO users (phone, name, role)
		VALUES ($1, $1, 'farmer')
		ON CONFLICT (phone) DO UPDATE SET last_login_at = NOW()
		RETURNING id, phone, name, role, language_pref`,
		phone,
	).Scan(&user.ID, &user.Phone, &user.Name, &user.Role, &user.Language)
	if err != nil {
		return "", nil, fmt.Errorf("upsert user: %w", err)
	}

	token, err := s.issueJWT(user)
	if err != nil {
		return "", nil, fmt.Errorf("issue jwt: %w", err)
	}

	return token, user, nil
}

// ValidateJWT validates a Bearer token and extracts claims.
func (s *AuthService) ValidateJWT(tokenStr string) (*User, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return &User{
		ID:    claims["sub"].(string),
		Phone: claims["phone"].(string),
		Name:  claims["name"].(string),
		Role:  claims["role"].(string),
	}, nil
}

func (s *AuthService) issueJWT(user *User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"phone": user.Phone,
		"name":  user.Name,
		"role":  user.Role,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}

func generateOTP(digits int) (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", digits, n), nil
}

func hashOTP(otp string) string {
	b := make([]byte, 16)
	copy(b, []byte(otp+"-storedge-salt"))
	return hex.EncodeToString(b)
}
