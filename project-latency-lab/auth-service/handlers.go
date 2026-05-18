package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

type RegisterRequest struct {
	Username string `json:"username"`
}

type RegisterResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type VerifyRequest struct {
	Token string `json:"token"`
}

type VerifyResponse struct {
	Username string `json:"username"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handleRegister(store *Store, cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "handler.Register")
		defer span.End()
		applyDelay(ctx, StageHandler)

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
			return
		}
		req.Username = strings.TrimSpace(req.Username)
		if req.Username == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "username required"})
			return
		}
		span.SetAttributes(attribute.String("username", req.Username))

		if shouldFail(ctx) {
			span.SetAttributes(attribute.Bool("fault.injected", true))
			logger(ctx).Warn("fault injected", "endpoint", "register", "fail_rate", optsFrom(ctx).failRate)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "injected failure"})
			return
		}

		token, err := generateToken()
		if err != nil {
			logger(ctx).Error("generate token failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to register"})
			return
		}
		if err := store.UpsertUser(ctx, req.Username, token); err != nil {
			logger(ctx).Error("upsert user failed", "username", req.Username, "error", err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to register"})
			return
		}
		cache.SetToken(ctx, token, req.Username)
		logger(ctx).Info("user registered", "username", req.Username)

		writeJSON(w, http.StatusCreated, RegisterResponse{Username: req.Username, Token: token})
	}
}

func handleVerify(store *Store, cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "handler.Verify")
		defer span.End()
		applyDelay(ctx, StageHandler)

		var req VerifyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
			return
		}
		if req.Token == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "token required"})
			return
		}

		if shouldFail(ctx) {
			span.SetAttributes(attribute.Bool("fault.injected", true))
			logger(ctx).Warn("fault injected", "endpoint", "verify", "fail_rate", optsFrom(ctx).failRate)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "injected failure"})
			return
		}

		username, err := cache.GetToken(ctx, req.Token)
		cacheHit := err == nil
		if err != nil {
			username, err = store.GetUsernameByToken(ctx, req.Token)
			if err != nil {
				if err == sql.ErrNoRows {
					logger(ctx).Info("token rejected", "reason", "not_found")
					writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
					return
				}
				logger(ctx).Error("verify token db error", "error", err)
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
				return
			}
			cache.SetToken(ctx, req.Token, username)
		}

		span.SetAttributes(attribute.String("username", username))
		logger(ctx).Debug("token verified", "username", username, "cache_hit", cacheHit)
		writeJSON(w, http.StatusOK, VerifyResponse{Username: username})
	}
}

func handleHealth(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := store.db.Ping(); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy", "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func generateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
