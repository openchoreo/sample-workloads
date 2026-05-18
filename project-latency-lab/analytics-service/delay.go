package main

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const (
	StageHandler = "handler"
	StageCache   = "cache"
	StageDB      = "db"
	StageAll     = "all"
)

type delayKey struct{}

type delayOpts struct {
	delayMs  int
	stage    string
	failRate float64
}

func delayMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		opts := delayOpts{stage: StageHandler}

		if v := q.Get("delay_ms"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				if n > 30000 {
					n = 30000
				}
				opts.delayMs = n
			}
		}
		if v := q.Get("delay_stage"); v != "" {
			switch v {
			case StageHandler, StageCache, StageDB, StageAll:
				opts.stage = v
			}
		}
		if v := q.Get("fail_rate"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
				if f > 1 {
					f = 1
				}
				opts.failRate = f
			}
		}

		ctx := context.WithValue(r.Context(), delayKey{}, opts)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func optsFrom(ctx context.Context) delayOpts {
	if v, ok := ctx.Value(delayKey{}).(delayOpts); ok {
		return v
	}
	return delayOpts{stage: StageHandler}
}

func applyDelay(ctx context.Context, stage string) {
	o := optsFrom(ctx)
	if o.delayMs <= 0 {
		return
	}
	if o.stage != StageAll && o.stage != stage {
		return
	}
	logger(ctx).Info("injecting latency", "stage", stage, "delay_ms", o.delayMs)
	t := time.NewTimer(time.Duration(o.delayMs) * time.Millisecond)
	defer t.Stop()
	select {
	case <-t.C:
	case <-ctx.Done():
	}
}

func shouldFail(ctx context.Context) bool {
	o := optsFrom(ctx)
	if o.failRate <= 0 {
		return false
	}
	return rand.Float64() < o.failRate
}
