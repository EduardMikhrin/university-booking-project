package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/cache"
	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/redis/go-redis/v9"
)

const (
	monthlyStatsListKey      = "reports:monthly:list"
	detailedMonthlyStatsPrefix = "reports:monthly:"
	reportsCachePattern      = "reports:*"
)

// ReportCache implements cache.ReportCacheQ interface using Redis
type ReportCache struct {
	client *redis.Client
}

// NewReportCache creates a new ReportCache instance
func NewReportCache(client *redis.Client) cache.ReportCacheQ {
	return &ReportCache{client: client}
}

// SetMonthlyStatsList caches list of monthly statistics
func (c *ReportCache) SetMonthlyStatsList(ctx context.Context, stats []*types.MonthlyStats, expiration time.Duration) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, monthlyStatsListKey, data, expiration).Err()
}

// GetMonthlyStatsList retrieves cached monthly statistics list
func (c *ReportCache) GetMonthlyStatsList(ctx context.Context) ([]*types.MonthlyStats, error) {
	val, err := c.client.Get(ctx, monthlyStatsListKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("monthly stats list not found in cache")
		}
		return nil, err
	}

	var stats []*types.MonthlyStats
	if err := json.Unmarshal([]byte(val), &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// SetDetailedMonthlyStats caches detailed monthly statistics
func (c *ReportCache) SetDetailedMonthlyStats(ctx context.Context, month string, stats *types.DetailedMonthlyStats, expiration time.Duration) error {
	key := detailedMonthlyStatsPrefix + month
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// GetDetailedMonthlyStats retrieves cached detailed monthly statistics
func (c *ReportCache) GetDetailedMonthlyStats(ctx context.Context, month string) (*types.DetailedMonthlyStats, error) {
	key := detailedMonthlyStatsPrefix + month
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("detailed monthly stats not found in cache")
		}
		return nil, err
	}

	var stats types.DetailedMonthlyStats
	if err := json.Unmarshal([]byte(val), &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// InvalidateMonthlyStats invalidates monthly statistics cache
func (c *ReportCache) InvalidateMonthlyStats(ctx context.Context, month string) error {
	key := detailedMonthlyStatsPrefix + month
	return c.client.Del(ctx, key).Err()
}

// InvalidateAllStats invalidates all statistics cache
func (c *ReportCache) InvalidateAllStats(ctx context.Context) error {
	// Delete all report keys using pattern matching
	iter := c.client.Scan(ctx, 0, reportsCachePattern, 0).Iterator()
	keys := []string{}
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

