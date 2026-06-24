package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/storedge/storedge/services/search-match/internal/matching"
	"github.com/storedge/storedge/services/search-match/internal/service"
)

type WarehouseHandler struct {
	matchingSvc *service.MatchingService
	logger      *zap.Logger
}

func NewWarehouseHandler(matchingSvc *service.MatchingService, logger *zap.Logger) *WarehouseHandler {
	return &WarehouseHandler{matchingSvc: matchingSvc, logger: logger}
}

// Search handles GET /api/v1/warehouses/search
// Query params: lat, lon, type, pallets, max_distance_km, max_price, wdra_only,
//               w_distance, w_price, w_quality, w_temperature, page, page_size
func (h *WarehouseHandler) Search(c *gin.Context) {
	lat := parseFloat(c.Query("lat"), 27.1767)   // Default: Agra
	lon := parseFloat(c.Query("lon"), 78.0081)
	pallets := parseInt(c.Query("pallets"), 1)
	maxDist := parseFloat(c.Query("max_distance_km"), 50)
	maxPrice := parseFloat(c.Query("max_price"), 0)
	wdraOnly := c.Query("wdra_only") == "true"
	page := parseInt(c.Query("page"), 0)
	pageSize := parseInt(c.Query("page_size"), 20)

	req := matching.SearchRequest{
		OriginLat:         lat,
		OriginLon:         lon,
		WarehouseType:     c.Query("type"),
		RequiredPallets:   pallets,
		MaxDistanceKm:     maxDist,
		MaxPricePerPallet: maxPrice,
		WDRAOnly:          wdraOnly,
		WeightDistance:    parseFloat(c.Query("w_distance"), 0.25),
		WeightPrice:       parseFloat(c.Query("w_price"), 0.25),
		WeightQuality:     parseFloat(c.Query("w_quality"), 0.25),
		WeightTemperature: parseFloat(c.Query("w_temperature"), 0.25),
		Page:              page,
		PageSize:          pageSize,
	}

	// Cold chain params
	if minTemp := c.Query("min_temp"); minTemp != "" {
		req.MinTemperature = parseFloat(minTemp, 0)
	}
	if maxTemp := c.Query("max_temp"); maxTemp != "" {
		req.MaxTemperature = parseFloat(maxTemp, 0)
	}

	results, total, err := h.matchingSvc.Search(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("search failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results":    results,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

// GetWarehouse handles GET /api/v1/warehouses/:id
func (h *WarehouseHandler) GetWarehouse(c *gin.Context) {
	id := c.Param("id")
	warehouse, err := h.matchingSvc.GetWarehouse(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch warehouse"})
		return
	}
	if warehouse == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "warehouse not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"warehouse": warehouse})
}

// GetDynamicPrice handles GET /api/v1/warehouses/:id/price
func (h *WarehouseHandler) GetDynamicPrice(c *gin.Context) {
	id := c.Param("id")
	warehouse, err := h.matchingSvc.GetWarehouse(c.Request.Context(), id)
	if err != nil || warehouse == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "warehouse not found"})
		return
	}

	occupancy := 1.0 - (float64(warehouse.AvailablePalletSlots) / float64(warehouse.TotalPalletCapacity))
	tier := "standard"
	if occupancy > 0.90 {
		tier = "peak"
	} else if occupancy < 0.60 {
		tier = "off-peak"
	}

	c.JSON(http.StatusOK, gin.H{
		"warehouse_id":               id,
		"price_per_pallet_inr":       warehouse.CurrentPrice,
		"occupancy_rate":             occupancy,
		"pricing_tier":               tier,
	})
}

func parseFloat(s string, defaultVal float64) float64 {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultVal
	}
	return v
}

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}
