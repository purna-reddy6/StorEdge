package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/storedge/storedge/services/search-match/internal/matching"
)

type WarehouseRepository struct {
	db *sql.DB
}

func NewWarehouseRepository(db *sql.DB) *WarehouseRepository {
	return &WarehouseRepository{db: db}
}

// SearchNearby queries warehouses within radius using PostGIS ST_DWithin.
// Applies type, temperature, WDRA, and availability filters.
func (r *WarehouseRepository) SearchNearby(ctx context.Context, req matching.SearchRequest) ([]matching.Warehouse, error) {
	query := `
		SELECT
			w.id,
			w.owner_id,
			w.name,
			w.description,
			w.type,
			w.wdra_status,
			COALESCE(w.wdra_registration_number, '') AS wdra_reg_number,
			ST_X(w.geo_location::geometry) AS longitude,
			ST_Y(w.geo_location::geometry) AS latitude,
			w.address_line1,
			w.city,
			w.state,
			w.pincode,
			w.total_pallet_capacity,
			w.available_pallet_slots,
			COALESCE(w.floor_area_sqft, 0) AS floor_area_sqft,
			COALESCE(
				w.available_pallet_slots::float / NULLIF(w.total_pallet_capacity, 0),
				0
			) AS occupancy_inverse,
			COALESCE(w.min_temperature_celsius, -99) AS min_temp,
			COALESCE(w.max_temperature_celsius, 99) AS max_temp,
			w.base_price_per_pallet_inr,
			COALESCE(w.current_dynamic_price_inr, w.base_price_per_pallet_inr) AS current_price,
			w.rating,
			w.total_reviews,
			w.apmc_license_number IS NOT NULL AS apmc_licensed,
			w.gst_number IS NOT NULL AS gst_registered,
			COALESCE(w.cover_image_url, '') AS cover_image_url,
			ST_Distance(
				w.geo_location::geography,
				ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
			) / 1000 AS distance_km
		FROM warehouses w
		WHERE
			w.is_active = TRUE
			AND w.status = 'active'
			AND w.available_pallet_slots >= $3
			AND ST_DWithin(
				w.geo_location::geography,
				ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
				$4 * 1000
			)
	`

	args := []interface{}{
		req.OriginLon,
		req.OriginLat,
		req.RequiredPallets,
		req.MaxDistanceKm,
	}
	argIdx := 5

	if req.WarehouseType != "" && req.WarehouseType != "any" {
		query += fmt.Sprintf(" AND w.type = $%d", argIdx)
		args = append(args, req.WarehouseType)
		argIdx++
	}

	if req.WDRAOnly {
		query += fmt.Sprintf(" AND w.wdra_status = $%d", argIdx)
		args = append(args, "registered")
		argIdx++
	}

	if req.MaxPricePerPallet > 0 {
		query += fmt.Sprintf(" AND COALESCE(w.current_dynamic_price_inr, w.base_price_per_pallet_inr) <= $%d", argIdx)
		args = append(args, req.MaxPricePerPallet)
		argIdx++
	}

	// Cold chain temperature compatibility
	if req.MinTemperature != 0 || req.MaxTemperature != 0 {
		query += fmt.Sprintf(` AND w.min_temperature_celsius <= $%d AND w.max_temperature_celsius >= $%d`, argIdx, argIdx+1)
		args = append(args, req.MinTemperature, req.MaxTemperature)
		argIdx += 2
	}

	query += " ORDER BY distance_km ASC LIMIT 100"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search nearby warehouses: %w", err)
	}
	defer rows.Close()

	var warehouses []matching.Warehouse
	for rows.Next() {
		var w matching.Warehouse
		err := rows.Scan(
			&w.ID, &w.OwnerID, &w.Name, &w.Description,
			&w.Type, &w.WDRAStatus, &w.WDRARegNumber,
			&w.Longitude, &w.Latitude,
			&w.AddressLine1, &w.City, &w.State, &w.Pincode,
			&w.TotalPalletCapacity, &w.AvailablePalletSlots, &w.FloorAreaSqft,
			&w.OccupancyInverse,
			&w.MinTemperature, &w.MaxTemperature,
			&w.BasePricePerPallet, &w.CurrentPrice,
			&w.Rating, &w.TotalReviews,
			&w.APMCLicensed, &w.GSTRegistered,
			&w.CoverImageURL,
			&w.DistanceKm,
		)
		if err != nil {
			return nil, fmt.Errorf("scan warehouse row: %w", err)
		}
		warehouses = append(warehouses, w)
	}
	return warehouses, rows.Err()
}

// GetByID fetches a single warehouse by ID.
func (r *WarehouseRepository) GetByID(ctx context.Context, id string) (*matching.Warehouse, error) {
	query := `
		SELECT
			w.id, w.owner_id, w.name, w.description, w.type, w.wdra_status,
			COALESCE(w.wdra_registration_number, ''),
			ST_X(w.geo_location::geometry), ST_Y(w.geo_location::geometry),
			w.address_line1, w.city, w.state, w.pincode,
			w.total_pallet_capacity, w.available_pallet_slots,
			COALESCE(w.floor_area_sqft, 0),
			COALESCE(w.available_pallet_slots::float / NULLIF(w.total_pallet_capacity, 0), 0),
			COALESCE(w.min_temperature_celsius, -99),
			COALESCE(w.max_temperature_celsius, 99),
			w.base_price_per_pallet_inr,
			COALESCE(w.current_dynamic_price_inr, w.base_price_per_pallet_inr),
			w.rating, w.total_reviews,
			w.apmc_license_number IS NOT NULL,
			w.gst_number IS NOT NULL,
			COALESCE(w.cover_image_url, ''),
			0 AS distance_km
		FROM warehouses w
		WHERE w.id = $1 AND w.is_active = TRUE`

	var w matching.Warehouse
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&w.ID, &w.OwnerID, &w.Name, &w.Description,
		&w.Type, &w.WDRAStatus, &w.WDRARegNumber,
		&w.Longitude, &w.Latitude,
		&w.AddressLine1, &w.City, &w.State, &w.Pincode,
		&w.TotalPalletCapacity, &w.AvailablePalletSlots, &w.FloorAreaSqft,
		&w.OccupancyInverse,
		&w.MinTemperature, &w.MaxTemperature,
		&w.BasePricePerPallet, &w.CurrentPrice,
		&w.Rating, &w.TotalReviews,
		&w.APMCLicensed, &w.GSTRegistered,
		&w.CoverImageURL,
		&w.DistanceKm,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get warehouse by id: %w", err)
	}
	return &w, nil
}

// UpdateDynamicPrice persists the AI-computed dynamic price to the warehouse.
func (r *WarehouseRepository) UpdateDynamicPrice(ctx context.Context, warehouseID string, price float64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE warehouses SET current_dynamic_price_inr = $1, updated_at = NOW() WHERE id = $2",
		price, warehouseID,
	)
	return err
}
