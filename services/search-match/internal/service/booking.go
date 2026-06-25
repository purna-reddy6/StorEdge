package service

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/storedge/storedge/services/search-match/internal/matching"
	"github.com/storedge/storedge/services/search-match/internal/repository"
)

const (
	CommissionRate = 0.10 // 10% platform fee
)

type BookingService struct {
	bookingRepo  *repository.BookingRepository
	warehouseRepo *repository.WarehouseRepository
	pricingCache *repository.PricingCache
	logger       *zap.Logger
}

func NewBookingService(
	bookingRepo *repository.BookingRepository,
	warehouseRepo *repository.WarehouseRepository,
	pricingCache *repository.PricingCache,
	logger *zap.Logger,
) *BookingService {
	return &BookingService{
		bookingRepo:  bookingRepo,
		warehouseRepo: warehouseRepo,
		pricingCache: pricingCache,
		logger:       logger,
	}
}

// CreateBooking validates the request, locks slots in Redis, computes pricing,
// and persists the booking atomically.
func (s *BookingService) CreateBooking(ctx context.Context, req matching.CreateBookingRequest) (*matching.Booking, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date: %w", err)
	}

	if !endDate.After(startDate) {
		return nil, fmt.Errorf("end_date must be after start_date")
	}

	warehouse, err := s.warehouseRepo.GetByID(ctx, req.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("fetch warehouse: %w", err)
	}
	if warehouse == nil {
		return nil, fmt.Errorf("warehouse not found")
	}
	if warehouse.AvailablePalletSlots < req.PalletCount {
		return nil, fmt.Errorf("insufficient capacity: %d pallets available, %d requested",
			warehouse.AvailablePalletSlots, req.PalletCount)
	}

	// Optimistic slot lock in Redis to prevent race conditions (TTL: 5 min)
	locked, err := s.pricingCache.LockSlots(ctx, req.WarehouseID, req.PalletCount, 5*time.Minute)
	if err != nil {
		s.logger.Warn("redis lock failed, continuing without cache lock", zap.Error(err))
	} else if !locked {
		return nil, fmt.Errorf("warehouse slots temporarily locked — please retry in a moment")
	}
	defer s.pricingCache.UnlockSlots(ctx, req.WarehouseID)

	// Get price (cached dynamic price or base price)
	price, ok := s.pricingCache.GetDynamicPrice(ctx, req.WarehouseID)
	if !ok {
		price = warehouse.CurrentPrice
	}

	// Calculate duration in months (pro-rated)
	durationDays := endDate.Sub(startDate).Hours() / 24
	durationMonths := durationDays / 30.0

	totalAmount := price * float64(req.PalletCount) * durationMonths
	commission := totalAmount * CommissionRate
	payout := totalAmount - commission

	booking := &matching.Booking{
		BookingNumber:    generateBookingNumber(),
		TenantID:         req.TenantID,
		WarehouseID:      req.WarehouseID,
		PalletCount:      req.PalletCount,
		CommodityType:    req.CommodityType,
		PricePerPallet:   price,
		TotalAmount:      totalAmount,
		CommissionAmount: commission,
		PayoutAmount:     payout,
		StartDate:        startDate,
		EndDate:          endDate,
		Status:           "pending",
	}

	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, fmt.Errorf("create booking: %w", err)
	}

	s.logger.Info("booking created",
		zap.String("booking_id", booking.ID),
		zap.String("booking_number", booking.BookingNumber),
		zap.Float64("total_amount_inr", totalAmount),
	)

	return booking, nil
}

// GetBooking fetches a booking by ID.
func (s *BookingService) GetBooking(ctx context.Context, id string) (*matching.Booking, error) {
	return s.bookingRepo.GetByID(ctx, id)
}

// ListBookings returns bookings for a user (all bookings for admins/operators when userID is "").
func (s *BookingService) ListBookings(ctx context.Context, userID string) ([]*matching.Booking, error) {
	return s.bookingRepo.ListByUser(ctx, userID)
}

func generateBookingNumber() string {
	year := time.Now().Year()
	random := rand.Intn(999999)
	return fmt.Sprintf("SE-%d-%06s", year, strconv.Itoa(random))
}
