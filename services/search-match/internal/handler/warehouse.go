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
	lat := parseFloat(firstOf(c.Query("lat")), 27.1767)
	// Accept both 'lng' (web/mobile) and 'lon' (legacy)
	lon := parseFloat(firstOf(c.Query("lng"), c.Query("lon")), 78.0081)
	pallets := parseInt(firstOf(c.Query("pallets")), 1)
	// Accept both 'radius_km' (web/mobile) and 'max_distance_km' (legacy)
	maxDist := parseFloat(firstOf(c.Query("radius_km"), c.Query("max_distance_km")), 50)
	maxPrice := parseFloat(firstOf(c.Query("max_price")), 0)
	wdraOnly := c.Query("wdra_only") == "true"
	page := parseInt(firstOf(c.Query("page")), 0)
	pageSize := parseInt(firstOf(c.Query("page_size")), 20)

	req := matching.SearchRequest{
		OriginLat:         lat,
		OriginLon:         lon,
		WarehouseType:     c.Query("type"),
		RequiredPallets:   pallets,
		MaxDistanceKm:     maxDist,
		MaxPricePerPallet: maxPrice,
		WDRAOnly:          wdraOnly,
		WeightDistance:    parseFloat(firstOf(c.Query("w_distance")), 0.25),
		WeightPrice:       parseFloat(firstOf(c.Query("w_price")), 0.25),
		WeightQuality:     parseFloat(firstOf(c.Query("w_quality")), 0.25),
		WeightTemperature: parseFloat(firstOf(c.Query("w_temperature")), 0.25),
		Page:              page,
		PageSize:          pageSize,
	}

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

	// Flatten: merge warehouse fields with match_score / distance_km at top level
	// so the frontend can access w.name, w.id, w.latitude directly.
	flat := make([]gin.H, 0, len(results))
	for _, r := range results {
		w := r.Warehouse
		entry := gin.H{
			"id": w.ID, "owner_id": w.OwnerID, "name": w.Name,
			"description": w.Description, "type": w.Type,
			"wdra_status": w.WDRAStatus, "wdra_registration_number": w.WDRARegNumber,
			"longitude": w.Longitude, "latitude": w.Latitude,
			"address_line1": w.AddressLine1, "city": w.City, "state": w.State, "pincode": w.Pincode,
			"total_pallet_capacity": w.TotalPalletCapacity, "available_pallet_slots": w.AvailablePalletSlots,
			"floor_area_sqft": w.FloorAreaSqft,
			"min_temperature_celsius": w.MinTemperature, "max_temperature_celsius": w.MaxTemperature,
			"price_per_pallet_per_day_inr": w.CurrentPrice,
			"base_price_per_pallet_inr": w.BasePricePerPallet,
			"rating": w.Rating, "total_reviews": w.TotalReviews,
			"apmc_licensed": w.APMCLicensed, "gst_registered": w.GSTRegistered,
			"cover_image_url": w.CoverImageURL,
			"distance_km": r.DistanceKm, "match_score": r.MatchScore,
			"estimated_monthly_cost_inr": r.EstimatedMonthlyCost,
		}
		flat = append(flat, entry)
	}

	c.JSON(http.StatusOK, gin.H{
		"warehouses": flat,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

// firstOf returns the first non-empty string from the arguments.
func firstOf(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
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
	c.JSON(http.StatusOK, warehouse)
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
