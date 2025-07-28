// Copyright 2025 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package repositories

import (
	"errors"
)

var ErrRecordNotFound = errors.New("record not found")
var ErrRecordAlreadyExists = errors.New("record already exists")
