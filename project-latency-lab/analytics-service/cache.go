package main

import (
	"context"
	"log"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
)

type Cache struct {
	client *redis.Client
}

func NewCache(addr string) *Cache {
	if addr == "" {
		return nil
	}
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		slog.Warn("redis unavailable, continuing without cache", "addr", addr, "error", err)
		return nil
	}
	log.Printf("redis connected at %s", addr)
	return &Cache{client: client}
}

// GetStats returns (count, recentViews, hit). hit=false means caller should
// fall through to the DB.
func (c *Cache) GetStats(ctx context.Context, id string) (int64, []string, bool) {
	if c == nil {
		return 0, nil, false
	}
	ctx, span := tracer.Start(ctx, "cache.GetStats")
	defer span.End()
	span.SetAttributes(attribute.String("note.id", id))
	applyDelay(ctx, StageCache)

	cs, err := c.client.Get(ctx, "views:"+id).Result()
	if err != nil {
		return 0, nil, false
	}
	count, err := strconv.ParseInt(cs, 10, 64)
	if err != nil {
		return 0, nil, false
	}
	recent, err := c.client.LRange(ctx, "recent:"+id, 0, 49).Result()
	if err != nil {
		return count, nil, true
	}
	return count, recent, true
}

func (c *Cache) SetStats(ctx context.Context, id string, count int64, recent []string) {
	if c == nil {
		return
	}
	ctx, span := tracer.Start(ctx, "cache.SetStats")
	defer span.End()
	span.SetAttributes(attribute.String("note.id", id))
	applyDelay(ctx, StageCache)

	pipe := c.client.Pipeline()
	pipe.Set(ctx, "views:"+id, count, 5*time.Minute)
	if len(recent) > 0 {
		pipe.Del(ctx, "recent:"+id)
		vals := make([]interface{}, len(recent))
		for i, v := range recent {
			vals[i] = v
		}
		pipe.RPush(ctx, "recent:"+id, vals...)
		pipe.Expire(ctx, "recent:"+id, 5*time.Minute)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		slog.Warn("redis set stats", "id", id, "error", err)
	}
}
