package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/cache"
	"github.com/EduardMikhrin/university-booking-project/internal/types"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	tableKeyPrefix            = "table:"
	tableNumberKeyPrefix      = "table:number:"
	allTablesKey              = "tables:all"
	availableTablesKeyPrefix  = "tables:available:"
	tableCachePattern         = "table:*"
	tablesCachePattern        = "tables:*"
)

// TableCache implements cache.TableCacheQ interface using Redis
type TableCache struct {
	client *redis.Client
}

// NewTableCache creates a new TableCache instance
func NewTableCache(client *redis.Client) cache.TableCacheQ {
	return &TableCache{client: client}
}

// SetTable caches a single table
func (c *TableCache) SetTable(ctx context.Context, tableID uuid.UUID, table *types.Table, expiration time.Duration) error {
	key := tableKeyPrefix + tableID.String()
	data, err := json.Marshal(table)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// GetTable retrieves cached table data
func (c *TableCache) GetTable(ctx context.Context, tableID uuid.UUID) (*types.Table, error) {
	key := tableKeyPrefix + tableID.String()
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("table not found in cache")
		}
		return nil, err
	}

	var table types.Table
	if err := json.Unmarshal([]byte(val), &table); err != nil {
		return nil, err
	}

	return &table, nil
}

// SetTableByNumber caches table by table number
func (c *TableCache) SetTableByNumber(ctx context.Context, number string, table *types.Table, expiration time.Duration) error {
	key := tableNumberKeyPrefix + number
	data, err := json.Marshal(table)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// GetTableByNumber retrieves cached table by number
func (c *TableCache) GetTableByNumber(ctx context.Context, number string) (*types.Table, error) {
	key := tableNumberKeyPrefix + number
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("table not found in cache")
		}
		return nil, err
	}

	var table types.Table
	if err := json.Unmarshal([]byte(val), &table); err != nil {
		return nil, err
	}

	return &table, nil
}

// SetAllTables caches list of all tables
func (c *TableCache) SetAllTables(ctx context.Context, tables []*types.Table, expiration time.Duration) error {
	data, err := json.Marshal(tables)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, allTablesKey, data, expiration).Err()
}

// GetAllTables retrieves cached list of all tables
func (c *TableCache) GetAllTables(ctx context.Context) ([]*types.Table, error) {
	val, err := c.client.Get(ctx, allTablesKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("tables not found in cache")
		}
		return nil, err
	}

	var tables []*types.Table
	if err := json.Unmarshal([]byte(val), &tables); err != nil {
		return nil, err
	}

	return tables, nil
}

// SetAvailableTables caches available tables for a specific date/time
func (c *TableCache) SetAvailableTables(ctx context.Context, date string, time string, guests int, tables []*types.Table, expiration time.Duration) error {
	key := fmt.Sprintf("%s%s:%s:%d", availableTablesKeyPrefix, date, time, guests)
	data, err := json.Marshal(tables)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// GetAvailableTables retrieves cached available tables
func (c *TableCache) GetAvailableTables(ctx context.Context, date string, time string, guests int) ([]*types.Table, error) {
	key := fmt.Sprintf("%s%s:%s:%d", availableTablesKeyPrefix, date, time, guests)
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("available tables not found in cache")
		}
		return nil, err
	}

	var tables []*types.Table
	if err := json.Unmarshal([]byte(val), &tables); err != nil {
		return nil, err
	}

	return tables, nil
}

// InvalidateTableCache invalidates all table-related cache
func (c *TableCache) InvalidateTableCache(ctx context.Context) error {
	// Delete all table keys using pattern matching
	iter := c.client.Scan(ctx, 0, tablesCachePattern, 0).Iterator()
	keys := []string{}
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}

	// Also add table: pattern
	iter = c.client.Scan(ctx, 0, tableCachePattern, 0).Iterator()
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

