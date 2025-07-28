// Copyright 2025 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package routes

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/wso2/choreo-sample-apps/go/rest-api/internal/models"

	"github.com/wso2/choreo-sample-apps/go/rest-api/internal/utils"
)

func registerReadingListRoutes(router fiber.Router) {
	r := router.Group("/reading-list/books")
	r.Post("/", AddBook)
	r.Get("/:id", GetBook)
	r.Put("/:id", UpdateBook)
	r.Delete("/:id", DeleteBook)
	r.Get("/", ListBooks)
}

// AddBook
//
//	@Summary	Add a new book to the reading list
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		request	body	models.Book	true	"New book details"
//	@Router		/books [post]
//	@Success	201	{object}	models.Book			"successful operation"
//	@Failure	400	{object}	utils.ErrorResponse	"invalid book details"
//	@Failure	409	{object}	utils.ErrorResponse	"book already exists"
func AddBook(c *fiber.Ctx) error {
	ctx := utils.GetRequestContext(c)
	newBook := models.Book{}
	if err := c.BodyParser(&newBook); err != nil {
		return makeHttpBadRequestError(err)
	}
	res, err := bookController.AddBook(ctx, newBook)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

// UpdateBook
//
//	@Summary	Update a reading list book by id
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		id		path	string		true	"Book ID"
//	@Param		request	body	models.Book	true	"Updated book details"
//	@Router		/books/{id} [put]
//	@Success	200	{object}	models.Book			"successful operation"
//	@Failure	400	{object}	utils.ErrorResponse	"invalid book details"
//	@Failure	404	{object}	utils.ErrorResponse	"book not found"
func UpdateBook(c *fiber.Ctx) error {
	ctx := utils.GetRequestContext(c)
	id := c.Params("id")
	updatedBook := models.Book{}
	if err := c.BodyParser(&updatedBook); err != nil {
		return makeHttpBadRequestError(err)
	}
	updatedBook.Id = id
	book, err := bookController.UpdateBook(ctx, updatedBook)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(book)
}

// DeleteBook
//
//	@Summary	Delete a reading list book by id
//	@Tags		books
//	@Produce	json
//	@Param		id	path	string	true	"Book ID"
//	@Router		/books/{id} [delete]
//	@Success	200	{object}	models.Book			"successful operation"
//	@Failure	404	{object}	utils.ErrorResponse	"book not found"
func DeleteBook(c *fiber.Ctx) error {
	ctx := utils.GetRequestContext(c)
	id := c.Params("id")
	book, err := bookController.DeleteBook(ctx, id)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(book)
}

// GetBook
//
//	@Summary	Get reading list book by id
//
//	@Tags		books
//
//	@Produce	json
//	@Param		id	path	string	true	"Book ID"
//	@Router		/books/{id} [get]
//	@Success	200	{object}	models.Book			"successful operation"
//	@Failure	404	{object}	utils.ErrorResponse	"book not found"
func GetBook(c *fiber.Ctx) error {
	ctx := utils.GetRequestContext(c)
	id := c.Params("id")

	book, err := bookController.GetBook(ctx, id)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(book)
}

// ListBooks
//
//	@Summary	List all the reading list books
//	@Tags		books
//	@Produce	json
//	@Router		/books [get]
//	@Success	200	{array}	models.Book	"successful operation"
func ListBooks(c *fiber.Ctx) error {
	ctx := utils.GetRequestContext(c)
	books, err := bookController.ListBooks(ctx)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(books)
}

func makeHttpBadRequestError(err error) *fiber.Error {
	return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("failed to parse the payload: %s", err.Error()))
}
