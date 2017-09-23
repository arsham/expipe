// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"fmt"
	"time"
)

var (
	// ErrEmptyName is the error when the package name is empty.
	ErrEmptyName = fmt.Errorf("name cannot be empty")

	// ErrEmptyEndpoint is the error when the given endpoint is empty.
	ErrEmptyEndpoint = fmt.Errorf("endpoint cannot be empty")

	// ErrEmptyTypeName is the error when the type_name is an empty string.
	ErrEmptyTypeName = fmt.Errorf("type_name cannot be empty")

	// ErrBackoffExceeded is the error when the endpoint's absence has
	// exceeded the backoff value. It is not strictly an error, it is
	// however a pointer to an error in the past.
	ErrBackoffExceeded = fmt.Errorf("endpoint gone too long")

	// ErrPingNotCalled is the error if the caller calls the record without pinging.
	ErrPingNotCalled = fmt.Errorf("the caller forgot to ask me pinging")
)

// ErrInvalidEndpoint is the error when the endpoint is not a valid url
type ErrInvalidEndpoint string

func (e ErrInvalidEndpoint) Error() string {
	return fmt.Sprintf("invalid endpoint: %s", string(e))
}

// ErrEndpointNotAvailable is the error when the endpoint is not available.
type ErrEndpointNotAvailable struct {
	Endpoint string
	Err      error
}

func (e ErrEndpointNotAvailable) Error() string {
	return fmt.Sprintf("endpoint (%s) not available: %s", e.Endpoint, e.Err)
}

// ErrLowBackoffValue is the error when the backoff value is lower than 5
type ErrLowBackoffValue int64

func (e ErrLowBackoffValue) Error() string {
	return fmt.Sprintf("back off should be at least 5: %d", e)
}

// ErrLowInterval is the error when the interval is zero
type ErrLowInterval time.Duration

func (e ErrLowInterval) Error() string {
	return fmt.Sprintf("interval should not be 0: %d", e)
}

// ErrLowTimeout is the error when the interval is zero
type ErrLowTimeout time.Duration

func (e ErrLowTimeout) Error() string {
	return fmt.Sprintf("timeout should be more than 1 second: %d", e)
}
