// Copyright 2025 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package routes

import (
	"github.com/openchoreo/sample-workloads/service-go-reading-list/internal/config"
	"github.com/openchoreo/sample-workloads/service-go-reading-list/internal/controllers"
	"github.com/openchoreo/sample-workloads/service-go-reading-list/internal/repositories"
)

var bookController *controllers.BookController

func initControllers() {
	initialData := config.LoadInitialData()
	bookRepository := repositories.NewBookRepository(initialData.Books)
	bookController = controllers.NewBookController(bookRepository)
}
