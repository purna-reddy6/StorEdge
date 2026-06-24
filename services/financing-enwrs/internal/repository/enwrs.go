package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ENWRsReceipt struct {
	ID                    string    `json:"id"`
	ReceiptNumber         string    `json:"receipt_number"`
	PalletItemID          string    `json:"pallet_item_id"`
	WarehouseID           string    `json:"warehouse_id"`
	DepositorID           string    `json:"depositor_id"`
	CommodityType         string    `json:"commodity_type"`
	CommodityVariety      string    `json:"commodity_variety"`
	QuantityKg            float64   `json:"quantity_kg"`
	QualityGrade          string    `json:"quality_grade"`
	MarketValueINR        float64   `json:"market_value_inr"`
	CurrentMarketValueINR float64   `json:"current_market_value_inr"`
	LTVRatio              float64   `json:"ltv_ratio"`
	MaxLoanAmountINR      float64   `json:"max_loan_amount_inr"`
	Repository            string    `json:"repository"`
	RepositoryReceiptID   string    `json:"repository_receipt_id"`
	WDRARegNumber         string    `json:"wdra_registration_number"`
	Status                string    `json:"status"`
	IssuedAt              *time.Time `json:"issued_at"`
	ExpiryDate            string    `json:"expiry_date"`
	OriginalQuantityKg    float64   `json:"original_quantity_kg"`
	ReleasedQuantityKg    float64   `json:"released_quantity_kg"`
	CreatedAt             time.Time `json:"created_at"`
}

type ENWRsRepository struct {
	db *sql.DB
}

func NewENWRsRepository(db *sql.DB) *ENWRsRepository {
	return &ENWRsRepository{db: db}
}

func (r *ENWRsRepository) Create(ctx context.Context, receipt *ENWRsReceipt) error {
	query := `
		INSERT INTO enwrs_receipts (
			receipt_number, pallet_item_id, warehouse_id, depositor_id,
			commodity_type, commodity_variety, quantity_kg, quality_grade,
			market_value_inr, current_market_value_inr, ltv_ratio,
			repository, wdra_registration_number,
			status, expiry_date, original_quantity_kg
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9,$10,$11,$12,'draft',$13,$7)
		RETURNING id, created_at`

	return r.db.QueryRowContext(ctx, query,
		receipt.ReceiptNumber, receipt.PalletItemID, receipt.WarehouseID, receipt.DepositorID,
		receipt.CommodityType, receipt.CommodityVariety, receipt.QuantityKg, receipt.QualityGrade,
		receipt.MarketValueINR, receipt.LTVRatio,
		receipt.Repository, receipt.WDRARegNumber, receipt.ExpiryDate,
	).Scan(&receipt.ID, &receipt.CreatedAt)
}

func (r *ENWRsRepository) GetByID(ctx context.Context, id string) (*ENWRsReceipt, error) {
	var rec ENWRsReceipt
	err := r.db.QueryRowContext(ctx, `
		SELECT id, receipt_number, pallet_item_id, warehouse_id, depositor_id,
			commodity_type, COALESCE(commodity_variety,''), quantity_kg, COALESCE(quality_grade,''),
			market_value_inr, COALESCE(current_market_value_inr, market_value_inr),
			ltv_ratio, max_loan_amount_inr, repository,
			COALESCE(repository_receipt_id,''), COALESCE(wdra_registration_number,''),
			status, issued_at, expiry_date::text,
			original_quantity_kg, released_quantity_kg, created_at
		FROM enwrs_receipts WHERE id = $1`, id,
	).Scan(
		&rec.ID, &rec.ReceiptNumber, &rec.PalletItemID, &rec.WarehouseID, &rec.DepositorID,
		&rec.CommodityType, &rec.CommodityVariety, &rec.QuantityKg, &rec.QualityGrade,
		&rec.MarketValueINR, &rec.CurrentMarketValueINR,
		&rec.LTVRatio, &rec.MaxLoanAmountINR, &rec.Repository,
		&rec.RepositoryReceiptID, &rec.WDRARegNumber,
		&rec.Status, &rec.IssuedAt, &rec.ExpiryDate,
		&rec.OriginalQuantityKg, &rec.ReleasedQuantityKg, &rec.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &rec, err
}

func (r *ENWRsRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE enwrs_receipts SET status = $1, updated_at = NOW() WHERE id = $2",
		status, id,
	)
	return err
}

func (r *ENWRsRepository) Issue(ctx context.Context, id, repositoryReceiptID string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		"UPDATE enwrs_receipts SET status = 'issued', issued_at = $1, repository_receipt_id = $2, updated_at = NOW() WHERE id = $3",
		now, repositoryReceiptID, id,
	)
	return err
}

func (r *ENWRsRepository) ListByDepositor(ctx context.Context, depositorID string) ([]ENWRsReceipt, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, receipt_number, commodity_type, quantity_kg, market_value_inr,
			max_loan_amount_inr, status, expiry_date::text, created_at
		FROM enwrs_receipts WHERE depositor_id = $1 ORDER BY created_at DESC`, depositorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var receipts []ENWRsReceipt
	for rows.Next() {
		var r ENWRsReceipt
		if err := rows.Scan(&r.ID, &r.ReceiptNumber, &r.CommodityType, &r.QuantityKg,
			&r.MarketValueINR, &r.MaxLoanAmountINR, &r.Status, &r.ExpiryDate, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		receipts = append(receipts, r)
	}
	return receipts, rows.Err()
}
