package matching

import "time"

// Warehouse represents a storage facility with all matching-relevant fields.
type Warehouse struct {
	ID                  string  `json:"id"`
	OwnerID             string  `json:"owner_id"`
	Name                string  `json:"name"`
	Description         string  `json:"description"`
	Type                string  `json:"type"`
	WDRAStatus          string  `json:"wdra_status"`
	WDRARegNumber       string  `json:"wdra_registration_number"`
	Longitude           float64 `json:"longitude"`
	Latitude            float64 `json:"latitude"`
	AddressLine1        string  `json:"address_line1"`
	City                string  `json:"city"`
	State               string  `json:"state"`
	Pincode             string  `json:"pincode"`
	TotalPalletCapacity int     `json:"total_pallet_capacity"`
	AvailablePalletSlots int    `json:"available_pallet_slots"`
	FloorAreaSqft       float64 `json:"floor_area_sqft"`
	OccupancyInverse    float64 `json:"-"` // available/total — used in scoring
	MinTemperature      float64 `json:"min_temperature_celsius"`
	MaxTemperature      float64 `json:"max_temperature_celsius"`
	BasePricePerPallet  float64 `json:"base_price_per_pallet_inr"`
	CurrentPrice        float64 `json:"current_dynamic_price_inr"`
	Rating              float64 `json:"rating"`
	TotalReviews        int     `json:"total_reviews"`
	APMCLicensed        bool    `json:"apmc_licensed"`
	GSTRegistered       bool    `json:"gst_registered"`
	CoverImageURL       string  `json:"cover_image_url"`
	DistanceKm          float64 `json:"distance_km"`
}

// WarehouseMatch is a scored search result.
type WarehouseMatch struct {
	Warehouse            Warehouse `json:"warehouse"`
	MatchScore           float64   `json:"match_score"`           // S = w1×D + w2×P + w3×Q + w4×T
	DistanceKm           float64   `json:"distance_km"`
	EstimatedMonthlyCost float64   `json:"estimated_monthly_cost_inr"`
}

// SearchRequest encodes all parameters for warehouse discovery.
type SearchRequest struct {
	OriginLat        float64 `json:"origin_lat"`
	OriginLon        float64 `json:"origin_lon"`
	WarehouseType    string  `json:"warehouse_type"`
	RequiredPallets  int     `json:"required_pallets"`
	MinTemperature   float64 `json:"min_temperature"`
	MaxTemperature   float64 `json:"max_temperature"`
	MaxDistanceKm    float64 `json:"max_distance_km"`
	MaxPricePerPallet float64 `json:"max_price_per_pallet"`
	WDRAOnly         bool    `json:"wdra_only"`

	// Tenant priority weights (must reflect business priorities)
	// Default: equal weights for MVP
	WeightDistance    float64 `json:"weight_distance"`    // w1
	WeightPrice       float64 `json:"weight_price"`       // w2
	WeightQuality     float64 `json:"weight_quality"`     // w3
	WeightTemperature float64 `json:"weight_temperature"` // w4

	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// Booking represents a storage reservation.
type Booking struct {
	ID               string    `json:"id"`
	BookingNumber    string    `json:"booking_number"`
	TenantID         string    `json:"tenant_id"`
	WarehouseID      string    `json:"warehouse_id"`
	WarehouseName    string    `json:"warehouse_name,omitempty"`
	FarmerName       string    `json:"farmer_name,omitempty"`
	PalletCount      int       `json:"pallet_count"`
	CommodityType    string    `json:"commodity_type"`
	PricePerPallet   float64   `json:"price_per_pallet_inr"`
	TotalAmount      float64   `json:"total_amount_inr"`
	CommissionAmount float64   `json:"commission_amount_inr"`
	PayoutAmount     float64   `json:"payout_amount_inr"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

// CreateBookingRequest is the input for creating a new booking.
type CreateBookingRequest struct {
	TenantID      string `json:"tenant_id"`
	WarehouseID   string `json:"warehouse_id"   binding:"required"`
	PalletCount   int    `json:"pallet_count"   binding:"required,min=1"`
	CommodityType string `json:"commodity_type" binding:"required"`
	StartDate     string `json:"start_date"     binding:"required"` // "2026-07-01"
	EndDate       string `json:"end_date"       binding:"required"` // "2026-09-30"
}
