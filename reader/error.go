// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"fmt"
	"time"
)

// ErrEmptyName is the error when the package name is empty.
// ErrEmptyEndpoint is the error when the given endpoint is empty.
// ErrEmptyTypeName is the error when the type_name is an empty string.
// ErrBackoffExceeded is the error when the endpoint's absence has exceeded the
// backoff value. It is not strictly an error, it is however a pointer to an
// error in the past.
// ErrPingNotCalled is the error if the caller calls the record without pinging.
var (
	ErrEmptyName       = fmt.Errorf("name cannot be empty")
	ErrEmptyEndpoint   = fmt.Errorf("endpoint cannot be empty")
	ErrEmptyTypeName   = fmt.Errorf("type_name cannot be empty")
	ErrBackoffExceeded = fmt.Errorf("endpoint gone too long")
	ErrPingNotCalled   = fmt.Errorf("the caller forgot to ask me pinging")
)

// InvalidEndpointError is the error when the endpoint is not a valid url.
type InvalidEndpointError string

func (e InvalidEndpointError) Error() string {
	return fmt.Sprintf("invalid endpoint: %s", string(e))
}

// EndpointNotAvailableError is the error when the endpoint is not available.
type EndpointNotAvailableError struct {
	Endpoint string
	Err      error
}

func (e EndpointNotAvailableError) Error() string {
	return fmt.Sprintf("endpoint (%s) not available: %s", e.Endpoint, e.Err)
}

// LowBackoffValueError is the error when the backoff value is lower than 5
type LowBackoffValueError int64

func (e LowBackoffValueError) Error() string {
	return fmt.Sprintf("back off should be at least 5: %d", e)
}

// LowIntervalError is the error when the interval is zero
type LowIntervalError time.Duration

func (e LowIntervalError) Error() string {
	return fmt.Sprintf("interval should not be 0: %d", e)
}

// LowTimeoutError is the error when the interval is zero
type LowTimeoutError time.Duration

func (e LowTimeoutError) Error() string {
	return fmt.Sprintf("timeout should be more than 1 second: %d", e)
}
