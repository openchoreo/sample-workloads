// Copyright 2025 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

const correlationIdHeaderName = "x-correlation-id"
const correlationIdCtxKey = "correlation-id"

func GetRequestContext(rCtx *fiber.Ctx) context.Context {
	correlationId := rCtx.Get(correlationIdHeaderName)
	ctx := context.Background()
	return context.WithValue(ctx, correlationIdCtxKey, correlationId)
}

func FiberErrorHandler(c *fiber.Ctx, err error) error {
	// Default 500 status code
	code := fiber.StatusInternalServerError
	msg := "internal server error"

	if e, ok := err.(*fiber.Error); ok {
		// Override status code if fiber.Error type
		code = e.Code
		msg = e.Error()
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	// Return status code with error message
	return c.Status(code).JSON(ErrorResponse{Message: msg})
}

type ErrorResponse struct {
	Message string `json:"message" example:"error message"`
}
