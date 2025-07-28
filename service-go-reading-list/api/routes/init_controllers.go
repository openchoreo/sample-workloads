// Copyright 2025 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package routes

import (
	"github.com/wso2/choreo-sample-apps/go/rest-api/internal/config"
	"github.com/wso2/choreo-sample-apps/go/rest-api/internal/controllers"
	"github.com/wso2/choreo-sample-apps/go/rest-api/internal/repositories"
)

var bookController *controllers.BookController

func initControllers() {
	initialData := config.LoadInitialData()
	bookRepository := repositories.NewBookRepository(initialData.Books)
	bookController = controllers.NewBookController(bookRepository)
}
