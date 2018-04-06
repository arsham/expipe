// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"fmt"
	"time"
)

// Errors regarding reading from an endpoint.
var (
	ErrEmptyName     = fmt.Errorf("name cannot be empty")
	ErrEmptyEndpoint = fmt.Errorf("endpoint cannot be empty")
	ErrEmptyTypeName = fmt.Errorf("type_name cannot be empty")
	ErrPingNotCalled = fmt.Errorf("the caller forgot to ask me pinging")
	ErrInvalidJSON   = fmt.Errorf("payload is invalid JSON object")
	ErrNillLogger    = fmt.Errorf("nil logger")
)

// InvalidEndpointError is the error when the endpoint is not a valid URL.
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

// LowIntervalError is the error when the interval is zero.
type LowIntervalError time.Duration

func (e LowIntervalError) Error() string {
	return fmt.Sprintf("interval should not be 0: %d", e)
}

// LowTimeoutError is the error when the interval is zero.
type LowTimeoutError time.Duration

func (e LowTimeoutError) Error() string {
	return fmt.Sprintf("timeout should be more than 1 second: %d", e)
}
