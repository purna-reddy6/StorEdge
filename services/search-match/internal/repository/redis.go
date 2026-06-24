package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedis(url string) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	rdb := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return rdb, nil
}

// PricingCache wraps Redis for pricing and inventory lock operations.
type PricingCache struct {
	rdb *redis.Client
}

func NewPricingCache(rdb *redis.Client) *PricingCache {
	return &PricingCache{rdb: rdb}
}

// GetDynamicPrice fetches the cached dynamic price for a warehouse.
func (c *PricingCache) GetDynamicPrice(ctx context.Context, warehouseID string) (float64, bool) {
	key := fmt.Sprintf("price:%s", warehouseID)
	val, err := c.rdb.Get(ctx, key).Float64()
	if err != nil {
		return 0, false
	}
	return val, true
}

// SetDynamicPrice caches the computed price for a warehouse (TTL: 15 minutes).
func (c *PricingCache) SetDynamicPrice(ctx context.Context, warehouseID string, price float64) error {
	key := fmt.Sprintf("price:%s", warehouseID)
	return c.rdb.Set(ctx, key, price, 15*time.Minute).Err()
}

// LockSlots acquires a temporary slot reservation lock to prevent double-booking.
// Returns true if lock acquired, false if slots already reserved.
func (c *PricingCache) LockSlots(ctx context.Context, warehouseID string, pallets int, ttl time.Duration) (bool, error) {
	key := fmt.Sprintf("lock:slots:%s", warehouseID)
	return c.rdb.SetNX(ctx, key, pallets, ttl).Result()
}

// UnlockSlots releases a slot reservation lock.
func (c *PricingCache) UnlockSlots(ctx context.Context, warehouseID string) error {
	key := fmt.Sprintf("lock:slots:%s", warehouseID)
	return c.rdb.Del(ctx, key).Err()
}

// GetOccupancyRate fetches cached occupancy rate for a warehouse.
func (c *PricingCache) GetOccupancyRate(ctx context.Context, warehouseID string) (float64, bool) {
	key := fmt.Sprintf("occupancy:%s", warehouseID)
	val, err := c.rdb.Get(ctx, key).Float64()
	if err != nil {
		return 0, false
	}
	return val, true
}

// SetOccupancyRate caches the occupancy rate (TTL: 5 minutes).
func (c *PricingCache) SetOccupancyRate(ctx context.Context, warehouseID string, rate float64) error {
	key := fmt.Sprintf("occupancy:%s", warehouseID)
	return c.rdb.Set(ctx, key, rate, 5*time.Minute).Err()
}
