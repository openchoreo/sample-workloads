// Copyright 2025 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"context"
)

type ReadStatus string

const (
	ReadStatusToRead  ReadStatus = "to_read"
	ReadStatusReading ReadStatus = "reading"
	ReadStatusRead    ReadStatus = "read"
)

func (s ReadStatus) String() string {
	return string(s)
}

type Book struct {
	Id     string     `json:"id" example:"fe2594d0-ccea-42a2-97ac-0487458b5642"`
	Title  string     `json:"title" example:"The Lord of the Rings"`
	Author string     `json:"author" example:"J. R. R. Tolkien"`
	Status ReadStatus `json:"status" example:"to_read" enums:"to_read,reading,read"`
}

type BookRepository interface {
	Add(ctx context.Context, book Book) (Book, error)
	Update(ctx context.Context, updatedBook Book) (Book, error)
	List(ctx context.Context) ([]Book, error)
	GetById(ctx context.Context, id string) (Book, error)
	DeleteById(ctx context.Context, id string) (Book, error)
}
