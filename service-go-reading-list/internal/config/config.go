// Copyright 2025 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/wso2/choreo-sample-apps/go/rest-api/internal/models"
)

type Config struct {
	// Env sets the environment the service is running in.
	// This is used in health check endpoint to indicate the environment.
	Env string
	// Hostname sets the hostname of the running service.
	// This is used to generate the Swagger host URL.
	Hostname string
	// Port sets the port of the running service.
	Port int
	// InitialDataPath sets the path to load the initial data file.
	// Refer to the InitialData struct for the file format.
	InitialDataPath string
}

type InitialData struct {
	Books []models.Book `json:"books"`
}
