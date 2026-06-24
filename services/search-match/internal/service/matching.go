package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"go.uber.org/zap"

	"github.com/storedge/storedge/services/search-match/internal/matching"
	"github.com/storedge/storedge/services/search-match/internal/repository"
)

type MatchingService struct {
	warehouseRepo *repository.WarehouseRepository
	pricingCache  *repository.PricingCache
	aiEngineURL   string
	httpClient    *http.Client
	logger        *zap.Logger
}

func NewMatchingService(
	warehouseRepo *repository.WarehouseRepository,
	pricingCache *repository.PricingCache,
	aiEngineURL string,
	logger *zap.Logger,
) *MatchingService {
	return &MatchingService{
		warehouseRepo: warehouseRepo,
		pricingCache:  pricingCache,
		aiEngineURL:   aiEngineURL,
		httpClient:    &http.Client{Timeout: 5 * time.Second},
		logger:        logger,
	}
}

// Search runs the full matching pipeline:
// 1. PostGIS spatial query for candidate warehouses
// 2. Dynamic pricing enrichment (cache → AI engine → base price)
// 3. Multi-dimensional scoring (blueprint formula: S = w1×D + w2×P + w3×Q + w4×T)
// 4. Sort by score descending, paginate
func (s *MatchingService) Search(ctx context.Context, req matching.SearchRequest) ([]matching.WarehouseMatch, int, error) {
	// Defaults
	if req.MaxDistanceKm == 0 {
		req.MaxDistanceKm = 50
	}
	if req.RequiredPallets == 0 {
		req.RequiredPallets = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	candidates, err := s.warehouseRepo.SearchNearby(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("search nearby: %w", err)
	}

	// Enrich each candidate with dynamic pricing
	for i := range candidates {
		candidates[i].CurrentPrice = s.resolvePrice(ctx, &candidates[i])
	}

	// Score and sort
	var matches []matching.WarehouseMatch
	for _, w := range candidates {
		score := matching.Score(w, req)
		months := candidates[0].DistanceKm // reuse distance slot — compute months
		_ = months
		durationDays := 30.0 // default 1 month for cost estimate
		matches = append(matches, matching.WarehouseMatch{
			Warehouse:            w,
			MatchScore:           score,
			DistanceKm:           w.DistanceKm,
			EstimatedMonthlyCost: w.CurrentPrice * float64(req.RequiredPallets) * durationDays / 30,
		})
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].MatchScore > matches[j].MatchScore
	})

	total := len(matches)
	start := (req.Page) * req.PageSize
	if start >= total {
		return []matching.WarehouseMatch{}, total, nil
	}
	end := start + req.PageSize
	if end > total {
		end = total
	}

	return matches[start:end], total, nil
}

// GetWarehouse fetches a single warehouse by ID.
func (s *MatchingService) GetWarehouse(ctx context.Context, id string) (*matching.Warehouse, error) {
	w, err := s.warehouseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, nil
	}
	w.CurrentPrice = s.resolvePrice(ctx, w)
	return w, nil
}

// resolvePrice gets the dynamic price: Redis cache → AI engine → base price fallback.
func (s *MatchingService) resolvePrice(ctx context.Context, w *matching.Warehouse) float64 {
	if price, ok := s.pricingCache.GetDynamicPrice(ctx, w.ID); ok {
		return price
	}

	price, err := s.fetchAIPrice(ctx, w)
	if err != nil {
		s.logger.Warn("AI pricing unavailable, using base price",
			zap.String("warehouse_id", w.ID),
			zap.Error(err),
		)
		return w.BasePricePerPallet
	}

	_ = s.pricingCache.SetDynamicPrice(ctx, w.ID, price)
	return price
}

type aiPriceRequest struct {
	WarehouseID     string  `json:"warehouse_id"`
	OccupancyRate   float64 `json:"occupancy_rate"`
	BasePriceINR    float64 `json:"base_price_inr"`
}

type aiPriceResponse struct {
	DynamicPrice float64 `json:"dynamic_price_inr"`
}

// fetchAIPrice calls the Python AI engine for a computed dynamic price.
// Blueprint formula: P = P_base × (1 + α(U - U*) + β×V)
func (s *MatchingService) fetchAIPrice(ctx context.Context, w *matching.Warehouse) (float64, error) {
	occupancyRate := 1.0 - w.OccupancyInverse // convert available fraction to occupancy fraction

	payload := aiPriceRequest{
		WarehouseID:   w.ID,
		OccupancyRate: occupancyRate,
		BasePriceINR:  w.BasePricePerPallet,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	url := fmt.Sprintf("%s/api/v1/pricing/compute", s.aiEngineURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url,
		io.NopCloser(stringReader(string(body))))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("AI engine returned status %d", resp.StatusCode)
	}

	var result aiPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.DynamicPrice, nil
}

type stringReaderType struct {
	s   string
	pos int
}

func (r *stringReaderType) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.s) {
		return 0, io.EOF
	}
	n = copy(p, r.s[r.pos:])
	r.pos += n
	return
}

func stringReader(s string) io.Reader {
	return &stringReaderType{s: s}
}
