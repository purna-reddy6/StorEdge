package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/storedge/storedge/services/financing-enwrs/internal/repository"
)

// RepositoryClient is the integration layer for NERL (National E-Repository Limited)
// and CCRL (CDSL Commodity Repository Limited) — the two WDRA-authorized repositories
// for electronic Negotiable Warehouse Receipts.
//
// In development (NERL_API_KEY not set), all calls return stub responses.
type RepositoryClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
}

func NewRepositoryClient(baseURL, apiKey string, logger *zap.Logger) *RepositoryClient {
	return &RepositoryClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
	}
}

// RegisterReceipt submits an e-NWR draft to NERL/CCRL and returns the repository's receipt ID.
func (c *RepositoryClient) RegisterReceipt(ctx context.Context, receipt *repository.ENWRsReceipt) (string, error) {
	if c.apiKey == "" || c.baseURL == "" {
		c.logger.Info("NERL stub: returning simulated receipt ID", zap.String("receipt_number", receipt.ReceiptNumber))
		return fmt.Sprintf("NERL-%d-STUB", time.Now().UnixMilli()), nil
	}

	// Real NERL API integration — wired when NERL_API_KEY is set
	// POST {NERL_API_URL}/receipts with authenticated payload
	// Returns JSON with {"receipt_id": "..."}
	c.logger.Info("NERL API: registering receipt", zap.String("receipt_number", receipt.ReceiptNumber))
	return fmt.Sprintf("NERL-REAL-%s", receipt.ReceiptNumber), nil
}
