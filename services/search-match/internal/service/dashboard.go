package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// DashboardService provides operator dashboard, inventory, and IoT alert data.
// It queries the shared PostgreSQL DB directly — no inter-service HTTP calls needed.
type DashboardService struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewDashboardService(db *sql.DB, logger *zap.Logger) *DashboardService {
	return &DashboardService{db: db, logger: logger}
}

// ─── Operator dashboard ──────────────────────────────────────────────────────

type OccupancyStat struct {
	Date         string  `json:"date"`
	OccupancyPct float64 `json:"occupancyPct"`
	Revenue      float64 `json:"revenue"`
}

// GetOccupancyStats returns daily occupancy + revenue for the past 30 days.
func (s *DashboardService) GetOccupancyStats(ctx context.Context, warehouseOwnerID string) ([]OccupancyStat, error) {
	query := `
		SELECT
			DATE(b.created_at)               AS day,
			AVG(1.0 - (w.available_pallet_slots::float / NULLIF(w.total_pallet_capacity,0))) AS occupancy,
			COALESCE(SUM(b.payout_amount_inr), 0) AS revenue
		FROM bookings b
		JOIN warehouses w ON w.id = b.warehouse_id
		WHERE b.created_at >= NOW() - INTERVAL '30 days'
		  AND ($1 = '' OR w.owner_id::text = $1)
		GROUP BY DATE(b.created_at)
		ORDER BY day`

	ownerArg := warehouseOwnerID
	if ownerArg == "" {
		ownerArg = ""
	}

	rows, err := s.db.QueryContext(ctx, query, ownerArg)
	if err != nil {
		return nil, fmt.Errorf("occupancy stats: %w", err)
	}
	defer rows.Close()

	var stats []OccupancyStat
	for rows.Next() {
		var st OccupancyStat
		var day time.Time
		if err := rows.Scan(&day, &st.OccupancyPct, &st.Revenue); err != nil {
			return nil, err
		}
		st.Date = day.Format("2006-01-02")
		stats = append(stats, st)
	}
	return stats, rows.Err()
}

// ─── IoT Alerts ──────────────────────────────────────────────────────────────

type IoTAlert struct {
	ID          string     `json:"id"`
	WarehouseID string     `json:"warehouse_id"`
	SensorID    string     `json:"sensor_id"`
	AlertType   string     `json:"alert_type"`
	Severity    string     `json:"severity"`
	Message     string     `json:"message"`
	IsResolved  bool       `json:"is_resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (s *DashboardService) ListAlerts(ctx context.Context, warehouseOwnerID string) ([]IoTAlert, error) {
	query := `
		SELECT a.id, a.warehouse_id, COALESCE(a.sensor_id::text, ''), a.alert_type,
			a.severity::text, a.message, a.is_resolved, a.resolved_at, a.triggered_at
		FROM iot_alerts a
		JOIN warehouses w ON w.id = a.warehouse_id
		WHERE ($1 = '' OR w.owner_id::text = $1)
		ORDER BY a.triggered_at DESC
		LIMIT 50`

	rows, err := s.db.QueryContext(ctx, query, warehouseOwnerID)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	defer rows.Close()

	var alerts []IoTAlert
	for rows.Next() {
		var a IoTAlert
		if err := rows.Scan(&a.ID, &a.WarehouseID, &a.SensorID, &a.AlertType,
			&a.Severity, &a.Message, &a.IsResolved, &a.ResolvedAt, &a.CreatedAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

func (s *DashboardService) ResolveAlert(ctx context.Context, alertID string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE iot_alerts SET is_resolved = TRUE, resolved_at = NOW() WHERE id = $1::uuid",
		alertID)
	return err
}

// ─── Inventory / Pallets ─────────────────────────────────────────────────────

type PalletItem struct {
	ID                        string     `json:"id"`
	BookingID                 string     `json:"bookingId"`
	Commodity                 string     `json:"commodity"`
	WeightKg                  float64    `json:"weightKg"`
	SlotPosition              string     `json:"slotPosition"`
	InwardDate                *time.Time `json:"inwardDate,omitempty"`
	ExpectedOutwardDate       *string    `json:"expectedOutwardDate,omitempty"`
	CurrentTemperatureCelsius *float64   `json:"currentTemperatureCelsius,omitempty"`
	ReleaseStatus             string     `json:"releaseStatus,omitempty"`
}

func (s *DashboardService) ListPallets(ctx context.Context, tenantID string) ([]PalletItem, error) {
	query := `
		SELECT p.id, p.booking_id, p.commodity_type,
			p.weight_kg, COALESCE(p.slot_position, ''), p.inward_date,
			p.expected_outward_date::text,
			(SELECT sr.temperature_celsius FROM sensor_readings sr
			 JOIN iot_sensors sens ON sens.id = sr.sensor_id
			 WHERE sens.warehouse_id = p.warehouse_id AND sr.temperature_celsius IS NOT NULL
			 ORDER BY sr.recorded_at DESC LIMIT 1) AS temp,
			COALESCE(srr.status::text, '') AS release_status
		FROM pallet_items p
		LEFT JOIN stock_release_requests srr ON srr.pallet_item_id = p.id
			AND srr.status NOT IN ('completed', 'rejected')
		WHERE ($1 = '' OR p.tenant_id::text = $1)
		ORDER BY p.inward_date DESC NULLS LAST
		LIMIT 100`

	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list pallets: %w", err)
	}
	defer rows.Close()

	var pallets []PalletItem
	for rows.Next() {
		var p PalletItem
		if err := rows.Scan(
			&p.ID, &p.BookingID, &p.Commodity,
			&p.WeightKg, &p.SlotPosition, &p.InwardDate, &p.ExpectedOutwardDate,
			&p.CurrentTemperatureCelsius, &p.ReleaseStatus,
		); err != nil {
			return nil, err
		}
		pallets = append(pallets, p)
	}
	return pallets, rows.Err()
}

func (s *DashboardService) InitiateRelease(ctx context.Context, palletItemID, requestedBy string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO stock_release_requests (id, pallet_item_id, requested_by_id, release_status)
		VALUES (uuid_generate_v4(), $1::uuid, $2::uuid, 'pending_otp')
		ON CONFLICT DO NOTHING`,
		palletItemID, requestedBy)
	return err
}

func (s *DashboardService) AuthorizeRelease(ctx context.Context, palletItemID, otp string) error {
	// In production this validates against otp_requests table.
	// For Phase 1 MVP: accept any 6-digit OTP to demonstrate the flow.
	if len(otp) != 6 {
		return fmt.Errorf("invalid OTP length")
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE stock_release_requests
		SET release_status = 'authorized', authorized_at = NOW()
		WHERE pallet_item_id = $1::uuid
		  AND release_status IN ('pending_otp', 'otp_sent')`,
		palletItemID)
	return err
}

