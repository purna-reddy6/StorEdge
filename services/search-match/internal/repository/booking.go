package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/storedge/storedge/services/search-match/internal/matching"
)

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// Create inserts a new booking and decrements available_pallet_slots atomically.
func (r *BookingRepository) Create(ctx context.Context, b *matching.Booking) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Lock the warehouse row and verify slots are still available
	var availableSlots int
	err = tx.QueryRowContext(ctx,
		"SELECT available_pallet_slots FROM warehouses WHERE id = $1 FOR UPDATE",
		b.WarehouseID,
	).Scan(&availableSlots)
	if err != nil {
		return fmt.Errorf("lock warehouse: %w", err)
	}

	if availableSlots < b.PalletCount {
		return fmt.Errorf("insufficient available slots: requested %d, available %d", b.PalletCount, availableSlots)
	}

	// Insert the booking
	query := `
		INSERT INTO bookings (
			id, booking_number, tenant_id, warehouse_id,
			pallet_count, commodity_type, price_per_pallet_inr,
			total_amount_inr, commission_amount_inr, payout_amount_inr,
			start_date, end_date, status
		) VALUES (
			uuid_generate_v4(), $1, $2, $3,
			$4, $5, $6,
			$7, $8, $9,
			$10, $11, 'pending'
		) RETURNING id, created_at`

	err = tx.QueryRowContext(ctx, query,
		b.BookingNumber, b.TenantID, b.WarehouseID,
		b.PalletCount, b.CommodityType, b.PricePerPallet,
		b.TotalAmount, b.CommissionAmount, b.PayoutAmount,
		b.StartDate, b.EndDate,
	).Scan(&b.ID, &b.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert booking: %w", err)
	}

	// Decrement available slots
	_, err = tx.ExecContext(ctx,
		"UPDATE warehouses SET available_pallet_slots = available_pallet_slots - $1, updated_at = NOW() WHERE id = $2",
		b.PalletCount, b.WarehouseID,
	)
	if err != nil {
		return fmt.Errorf("update warehouse slots: %w", err)
	}

	return tx.Commit()
}

// GetByID fetches a booking by its UUID.
func (r *BookingRepository) GetByID(ctx context.Context, id string) (*matching.Booking, error) {
	query := `
		SELECT id, booking_number, tenant_id, warehouse_id,
			pallet_count, commodity_type, price_per_pallet_inr,
			total_amount_inr, commission_amount_inr, payout_amount_inr,
			start_date, end_date, status, created_at
		FROM bookings WHERE id = $1`

	var b matching.Booking
	var startDate, endDate time.Time
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&b.ID, &b.BookingNumber, &b.TenantID, &b.WarehouseID,
		&b.PalletCount, &b.CommodityType, &b.PricePerPallet,
		&b.TotalAmount, &b.CommissionAmount, &b.PayoutAmount,
		&startDate, &endDate, &b.Status, &b.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get booking: %w", err)
	}
	b.StartDate = startDate
	b.EndDate = endDate
	return &b, nil
}

// UpdateStatus changes a booking's status (e.g., pending → confirmed).
func (r *BookingRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE bookings SET status = $1, updated_at = NOW() WHERE id = $2",
		status, id,
	)
	return err
}
