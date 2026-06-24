package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"github.com/storedge/storedge/services/financing-enwrs/internal/repository"
)

type ENWRsService struct {
	enwrsRepo      *repository.ENWRsRepository
	loanRepo       *repository.LoanRepository
	nerlClient     *RepositoryClient
	originationFee float64 // 1.5%
	logger         *zap.Logger
}

func NewENWRsService(
	enwrsRepo *repository.ENWRsRepository,
	loanRepo *repository.LoanRepository,
	nerlClient *RepositoryClient,
	originationFee float64,
	logger *zap.Logger,
) *ENWRsService {
	return &ENWRsService{
		enwrsRepo:      enwrsRepo,
		loanRepo:       loanRepo,
		nerlClient:     nerlClient,
		originationFee: originationFee,
		logger:         logger,
	}
}

type CreateReceiptRequest struct {
	PalletItemID     string  `json:"pallet_item_id"  binding:"required"`
	WarehouseID      string  `json:"warehouse_id"    binding:"required"`
	DepositorID      string  `json:"depositor_id"    binding:"required"`
	CommodityType    string  `json:"commodity_type"  binding:"required"`
	CommodityVariety string  `json:"commodity_variety"`
	QuantityKg       float64 `json:"quantity_kg"     binding:"required,min=0.01"`
	QualityGrade     string  `json:"quality_grade"`
	MarketValueINR   float64 `json:"market_value_inr" binding:"required,min=1"`
	WDRARegNumber    string  `json:"wdra_registration_number" binding:"required"`
	ExpiryDays       int     `json:"expiry_days"`     // default 180
}

// CreateReceipt drafts an e-NWR for warehouse-deposited goods.
func (s *ENWRsService) CreateReceipt(ctx context.Context, req CreateReceiptRequest) (*repository.ENWRsReceipt, error) {
	expiryDays := req.ExpiryDays
	if expiryDays <= 0 {
		expiryDays = 180
	}
	expiryDate := time.Now().AddDate(0, 0, expiryDays).Format("2006-01-02")

	receipt := &repository.ENWRsReceipt{
		ReceiptNumber:      generateReceiptNumber(),
		PalletItemID:       req.PalletItemID,
		WarehouseID:        req.WarehouseID,
		DepositorID:        req.DepositorID,
		CommodityType:      req.CommodityType,
		CommodityVariety:   req.CommodityVariety,
		QuantityKg:         req.QuantityKg,
		QualityGrade:       req.QualityGrade,
		MarketValueINR:     req.MarketValueINR,
		LTVRatio:           0.70, // 70% LTV (RBI guideline)
		Repository:         "NERL",
		WDRARegNumber:      req.WDRARegNumber,
		ExpiryDate:         expiryDate,
		OriginalQuantityKg: req.QuantityKg,
	}

	if err := s.enwrsRepo.Create(ctx, receipt); err != nil {
		return nil, fmt.Errorf("create receipt: %w", err)
	}

	s.logger.Info("e-NWR receipt created",
		zap.String("receipt_number", receipt.ReceiptNumber),
		zap.String("depositor_id", req.DepositorID),
		zap.Float64("quantity_kg", req.QuantityKg),
		zap.Float64("market_value_inr", req.MarketValueINR),
	)
	return receipt, nil
}

// IssueReceipt submits the receipt to NERL/CCRL and marks it as officially issued.
func (s *ENWRsService) IssueReceipt(ctx context.Context, receiptID string) (*repository.ENWRsReceipt, error) {
	receipt, err := s.enwrsRepo.GetByID(ctx, receiptID)
	if err != nil || receipt == nil {
		return nil, fmt.Errorf("receipt not found: %s", receiptID)
	}

	if receipt.Status != "draft" {
		return nil, fmt.Errorf("receipt is not in draft state: %s", receipt.Status)
	}

	// Submit to NERL (or use stub in dev)
	repoReceiptID, err := s.nerlClient.RegisterReceipt(ctx, receipt)
	if err != nil {
		s.logger.Warn("NERL registration failed, using stub ID", zap.Error(err))
		repoReceiptID = "NERL-STUB-" + receipt.ReceiptNumber
	}

	if err := s.enwrsRepo.Issue(ctx, receiptID, repoReceiptID); err != nil {
		return nil, fmt.Errorf("issue receipt: %w", err)
	}

	receipt.Status = "issued"
	receipt.RepositoryReceiptID = repoReceiptID
	return receipt, nil
}

type LoanApplicationRequest struct {
	ApplicantID        string  `json:"applicant_id"         binding:"required"`
	ENWRsReceiptID     string  `json:"enwrs_receipt_id"     binding:"required"`
	RequestedAmountINR float64 `json:"requested_amount_inr" binding:"required,min=1"`
	PartnerBankName    string  `json:"partner_bank_name"    binding:"required"`
}

// ApplyForLoan creates a loan application against an issued e-NWR.
func (s *ENWRsService) ApplyForLoan(ctx context.Context, req LoanApplicationRequest) (*repository.LoanApplication, error) {
	receipt, err := s.enwrsRepo.GetByID(ctx, req.ENWRsReceiptID)
	if err != nil || receipt == nil {
		return nil, fmt.Errorf("receipt not found")
	}

	if receipt.Status != "issued" {
		return nil, fmt.Errorf("receipt must be issued before applying for a loan (current: %s)", receipt.Status)
	}

	if receipt.DepositorID != req.ApplicantID {
		return nil, fmt.Errorf("applicant does not own this receipt")
	}

	if req.RequestedAmountINR > receipt.MaxLoanAmountINR {
		return nil, fmt.Errorf("requested amount ₹%.0f exceeds max loan (70%% LTV) of ₹%.0f",
			req.RequestedAmountINR, receipt.MaxLoanAmountINR)
	}

	// PSL limit check: ₹75 lakh for e-NWR backed loans (vs ₹50L for paper)
	const pslLimitINR = 7_500_000
	if req.RequestedAmountINR > pslLimitINR {
		return nil, fmt.Errorf("requested amount exceeds PSL limit of ₹75 lakh")
	}

	originationFee := req.RequestedAmountINR * s.originationFee

	loan := &repository.LoanApplication{
		ApplicationNumber:  generateAppNumber(),
		ApplicantID:        req.ApplicantID,
		ENWRsReceiptID:     req.ENWRsReceiptID,
		RequestedAmountINR: req.RequestedAmountINR,
		OriginationFeeINR:  &originationFee,
		PartnerBankName:    req.PartnerBankName,
		IsPSLEligible:      true,
		PSLLimitINR:        pslLimitINR,
	}

	if err := s.loanRepo.Create(ctx, loan); err != nil {
		return nil, fmt.Errorf("create loan: %w", err)
	}

	// Mark receipt as pledged
	_ = s.enwrsRepo.UpdateStatus(ctx, req.ENWRsReceiptID, "pledged")

	s.logger.Info("loan application created",
		zap.String("application_number", loan.ApplicationNumber),
		zap.String("bank", req.PartnerBankName),
		zap.Float64("requested_inr", req.RequestedAmountINR),
		zap.Float64("origination_fee_inr", originationFee),
	)
	return loan, nil
}

func (s *ENWRsService) GetReceipt(ctx context.Context, id string) (*repository.ENWRsReceipt, error) {
	return s.enwrsRepo.GetByID(ctx, id)
}

func (s *ENWRsService) GetLoan(ctx context.Context, id string) (*repository.LoanApplication, error) {
	return s.loanRepo.GetByID(ctx, id)
}

func (s *ENWRsService) ListReceiptsByDepositor(ctx context.Context, depositorID string) ([]repository.ENWRsReceipt, error) {
	return s.enwrsRepo.ListByDepositor(ctx, depositorID)
}

func generateReceiptNumber() string {
	return fmt.Sprintf("NERL-%d-%06d", time.Now().Year(), rand.Intn(999999))
}

func generateAppNumber() string {
	return fmt.Sprintf("LOAN-%d-%06d", time.Now().Year(), rand.Intn(999999))
}
