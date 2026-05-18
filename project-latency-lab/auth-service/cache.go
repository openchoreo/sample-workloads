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
	log.Printf("redis connected at %s", addr)
	return &Cache{client: client}
}

func (c *Cache) GetToken(ctx context.Context, token string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("cache disabled")
	}
	ctx, span := tracer.Start(ctx, "cache.GetToken")
	defer span.End()
	applyDelay(ctx, StageCache)

	v, err := c.client.Get(ctx, "tok:"+token).Result()
	if err == redis.Nil {
		span.SetAttributes(attribute.Bool("cache.hit", false))
		return "", fmt.Errorf("cache miss")
	}
	if err != nil {
		slog.Warn("redis get token", "error", err)
		return "", err
	}
	span.SetAttributes(attribute.Bool("cache.hit", true), attribute.String("username", v))
	return v, nil
}

func (c *Cache) SetToken(ctx context.Context, token, username string) {
	if c == nil {
		return
	}
	ctx, span := tracer.Start(ctx, "cache.SetToken")
	defer span.End()
	span.SetAttributes(attribute.String("username", username))
	applyDelay(ctx, StageCache)

	if err := c.client.Set(ctx, "tok:"+token, username, 1*time.Hour).Err(); err != nil {
		slog.Warn("redis set token", "error", err)
	}
}
