package repository

import (
	"context"
	"database/sql"
	"time"
)

type LoanApplication struct {
	ID                  string     `json:"id"`
	ApplicationNumber   string     `json:"application_number"`
	ApplicantID         string     `json:"applicant_id"`
	ENWRsReceiptID      string     `json:"enwrs_receipt_id"`
	RequestedAmountINR  float64    `json:"requested_amount_inr"`
	SanctionedAmountINR *float64   `json:"sanctioned_amount_inr"`
	InterestRatePercent *float64   `json:"interest_rate_percent"`
	TenureDays          *int       `json:"tenure_days"`
	OriginationFeeINR   *float64   `json:"origination_fee_inr"`
	PartnerBankName     string     `json:"partner_bank_name"`
	Status              string     `json:"status"`
	IssuedAmountINR     *float64   `json:"issued_amount_inr"`
	IsPSLEligible       bool       `json:"is_psl_eligible"`
	PSLLimitINR         float64    `json:"psl_limit_inr"`
	AppliedAt           time.Time  `json:"applied_at"`
	DisbursedAt         *time.Time `json:"disbursed_at"`
}

type LoanRepository struct {
	db *sql.DB
}

func NewLoanRepository(db *sql.DB) *LoanRepository {
	return &LoanRepository{db: db}
}

func (r *LoanRepository) Create(ctx context.Context, loan *LoanApplication) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO loan_applications (
			application_number, applicant_id, enwrs_receipt_id,
			requested_amount_inr, origination_fee_inr, partner_bank_name,
			status, is_psl_eligible, psl_limit_inr
		) VALUES ($1,$2,$3,$4,$5,$6,'applied',TRUE,7500000)
		RETURNING id, applied_at`,
		loan.ApplicationNumber, loan.ApplicantID, loan.ENWRsReceiptID,
		loan.RequestedAmountINR, loan.OriginationFeeINR, loan.PartnerBankName,
	).Scan(&loan.ID, &loan.AppliedAt)
}

func (r *LoanRepository) GetByID(ctx context.Context, id string) (*LoanApplication, error) {
	var loan LoanApplication
	err := r.db.QueryRowContext(ctx, `
		SELECT id, application_number, applicant_id, enwrs_receipt_id,
			requested_amount_inr, sanctioned_amount_inr, origination_fee_inr,
			COALESCE(partner_bank_name,''), status, is_psl_eligible, psl_limit_inr, applied_at
		FROM loan_applications WHERE id = $1`, id,
	).Scan(
		&loan.ID, &loan.ApplicationNumber, &loan.ApplicantID, &loan.ENWRsReceiptID,
		&loan.RequestedAmountINR, &loan.SanctionedAmountINR, &loan.OriginationFeeINR,
		&loan.PartnerBankName, &loan.Status, &loan.IsPSLEligible, &loan.PSLLimitINR, &loan.AppliedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &loan, err
}

func (r *LoanRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE loan_applications SET status = $1, updated_at = NOW() WHERE id = $2",
		status, id,
	)
	return err
}
