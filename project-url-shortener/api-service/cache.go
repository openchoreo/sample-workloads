package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
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
	log.Printf("Redis connected at %s", addr)
	return &Cache{client: client}
}

func (c *Cache) GetURL(ctx context.Context, shortCode string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("cache disabled")
	}
	ctx, span := tracer.Start(ctx, "cache.GetURL")
	defer span.End()
	span.SetAttributes(attribute.String("short_code", shortCode))

	val, err := c.client.Get(ctx, "url:"+shortCode).Result()
	if err == redis.Nil {
		span.SetAttributes(attribute.Bool("cache.hit", false))
		return "", fmt.Errorf("cache miss")
	}
	if err != nil {
		slog.Warn("failed to get URL from redis", "short_code", shortCode, "error", err)
		return "", err
	}
	span.SetAttributes(attribute.Bool("cache.hit", true))
	return val, nil
}

func (c *Cache) SetURL(ctx context.Context, shortCode, originalURL string) {
	if c == nil {
		return
	}
	ctx, span := tracer.Start(ctx, "cache.SetURL")
	defer span.End()

	if err := c.client.Set(ctx, "url:"+shortCode, originalURL, 1*time.Hour).Err(); err != nil {
		slog.Warn("failed to cache URL mapping in redis", "short_code", shortCode, "error", err)
	}
}

func (c *Cache) DeleteURL(ctx context.Context, shortCode string) {
	if c == nil {
		return
	}
	ctx, span := tracer.Start(ctx, "cache.DeleteURL")
	defer span.End()

	c.client.Del(ctx, "url:"+shortCode, "clicks:"+shortCode, "recent:"+shortCode)
}

func (c *Cache) IncrClickCount(ctx context.Context, shortCode string) {
	if c == nil {
		return
	}
	ctx, span := tracer.Start(ctx, "cache.IncrClickCount")
	defer span.End()

	pipe := c.client.Pipeline()
	pipe.Incr(ctx, "clicks:"+shortCode)
	pipe.Expire(ctx, "clicks:"+shortCode, 5*time.Minute)
	// Push to recent clicks list
	now := time.Now().UTC().Format(time.RFC3339)
	pipe.LPush(ctx, "recent:"+shortCode, now)
	pipe.LTrim(ctx, "recent:"+shortCode, 0, 49)
	pipe.Expire(ctx, "recent:"+shortCode, 5*time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		slog.Warn("failed to increment click count in redis", "short_code", shortCode, "error", err)
	}
}

func (c *Cache) GetClickCount(ctx context.Context, shortCode string) (int64, error) {
	if c == nil {
		return 0, fmt.Errorf("cache disabled")
	}
	ctx, span := tracer.Start(ctx, "cache.GetClickCount")
	defer span.End()

	val, err := c.client.Get(ctx, "clicks:"+shortCode).Int64()
	if err != nil {
		if err != redis.Nil {
			slog.Warn("failed to get click count from redis", "short_code", shortCode, "error", err)
		}
		return 0, err
	}
	return val, nil
}

func (c *Cache) GetRecentClicks(ctx context.Context, shortCode string) ([]string, error) {
	if c == nil {
		return nil, fmt.Errorf("cache disabled")
	}
	ctx, span := tracer.Start(ctx, "cache.GetRecentClicks")
	defer span.End()

	vals, err := c.client.LRange(ctx, "recent:"+shortCode, 0, 49).Result()
	if err != nil {
		slog.Warn("failed to get recent clicks from redis", "short_code", shortCode, "error", err)
	}
	return vals, err
}

func (c *Cache) SetClickData(ctx context.Context, shortCode string, count int64, recentClicks []string) {
	if c == nil {
		return
	}
	ctx, span := tracer.Start(ctx, "cache.SetClickData")
	defer span.End()

	pipe := c.client.Pipeline()
	pipe.Set(ctx, "clicks:"+shortCode, count, 5*time.Minute)
	if len(recentClicks) > 0 {
		pipe.Del(ctx, "recent:"+shortCode)
		vals := make([]interface{}, len(recentClicks))
		for i, v := range recentClicks {
			vals[i] = v
		}
		pipe.RPush(ctx, "recent:"+shortCode, vals...)
		pipe.Expire(ctx, "recent:"+shortCode, 5*time.Minute)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		slog.Warn("failed to cache click data in redis", "short_code", shortCode, "error", err)
	}
}
