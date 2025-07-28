// Copyright 2025 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package routes

import (
	"github.com/gofiber/fiber/v2"
)

func Initialize(app *fiber.App) {
	initControllers()

	RegisterHealthRoutes(app)
	apiVersion := app.Group("/api/v1")
	registerReadingListRoutes(apiVersion)
}
