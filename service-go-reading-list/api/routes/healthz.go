// Copyright 2025 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/wso2/choreo-sample-apps/go/rest-api/internal/config"
)

func HandleHealthCheckRequest(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"message":     "Reading list service is healthy",
		"environment": config.GetConfig().Env,
		"timestamp":   time.Now(),
	})
}

func RegisterHealthRoutes(r fiber.Router) {
	r.Get("/healthz", HandleHealthCheckRequest)
}
