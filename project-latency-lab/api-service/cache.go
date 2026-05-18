package main

import (
	"context"
	"encoding/json"
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

func (c *Cache) GetNote(ctx context.Context, id string) (*Note, error) {
	if c == nil {
		return nil, fmt.Errorf("cache disabled")
	}
	ctx, span := tracer.Start(ctx, "cache.GetNote")
	defer span.End()
	span.SetAttributes(attribute.String("note.id", id))
	applyDelay(ctx, StageCache)

	val, err := c.client.Get(ctx, "note:"+id).Bytes()
	if err == redis.Nil {
		span.SetAttributes(attribute.Bool("cache.hit", false))
		return nil, fmt.Errorf("cache miss")
	}
	span.SetAttributes(attribute.Bool("cache.hit", true))
	if err != nil {
		slog.Warn("redis get note", "id", id, "error", err)
		return nil, err
	}
	var n Note
	if err := json.Unmarshal(val, &n); err != nil {
		return nil, err
	}
	return &n, nil
}

func (c *Cache) SetNote(ctx context.Context, n *Note) {
	if c == nil || n == nil {
		return
	}
	ctx, span := tracer.Start(ctx, "cache.SetNote")
	defer span.End()
	span.SetAttributes(attribute.String("note.id", n.ID))
	applyDelay(ctx, StageCache)

	b, err := json.Marshal(n)
	if err != nil {
		return
	}
	if err := c.client.Set(ctx, "note:"+n.ID, b, 1*time.Hour).Err(); err != nil {
		slog.Warn("redis set note", "id", n.ID, "error", err)
	}
}

func (c *Cache) DeleteNote(ctx context.Context, id string) {
	if c == nil {
		return
	}
	ctx, span := tracer.Start(ctx, "cache.DeleteNote")
	defer span.End()
	span.SetAttributes(attribute.String("note.id", id))
	applyDelay(ctx, StageCache)
	c.client.Del(ctx, "note:"+id, "views:"+id, "recent:"+id)
}

func (c *Cache) IncrViewCount(ctx context.Context, id string) {
	if c == nil {
		return
	}
	ctx, span := tracer.Start(ctx, "cache.IncrViewCount")
	defer span.End()
	span.SetAttributes(attribute.String("note.id", id))
	applyDelay(ctx, StageCache)

	pipe := c.client.Pipeline()
	pipe.Incr(ctx, "views:"+id)
	pipe.Expire(ctx, "views:"+id, 5*time.Minute)
	now := time.Now().UTC().Format(time.RFC3339)
	pipe.LPush(ctx, "recent:"+id, now)
	pipe.LTrim(ctx, "recent:"+id, 0, 49)
	pipe.Expire(ctx, "recent:"+id, 5*time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		slog.Warn("redis incr views", "id", id, "error", err)
	}
}
