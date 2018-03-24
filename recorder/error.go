// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
	"fmt"
	"time"
)

// ErrEmptyName is the error when the package name is empty.
// ErrEmptyEndpoint is the error when the given endpoint is empty.
// ErrEmptyIndexName is the error when the index_name is an empty string.
// ErrBackoffExceeded is the error when the endpoint's absence has exceeded the
// backoff value. It is not strictly an error, it is however a pointer to an
// error in the past.
// ErrPingNotCalled is the error if the caller calls the record without pinging.
var (
	ErrEmptyName       = fmt.Errorf("name cannot be empty")
	ErrEmptyEndpoint   = fmt.Errorf("endpoint cannot be empty")
	ErrEmptyIndexName  = fmt.Errorf("index_name cannot be empty")
	ErrBackoffExceeded = fmt.Errorf("endpoint gone too long")
	ErrPingNotCalled   = fmt.Errorf("the caller forgot to ask me pinging")
)

// ErrInvalidEndpoint is the error when the endpoint is not a valid url
type ErrInvalidEndpoint string

func (e ErrInvalidEndpoint) Error() string { return fmt.Sprintf("invalid endpoint: %s", string(e)) }

// ErrLowBackoffValue is the error when the endpoint is not a valid url
type ErrLowBackoffValue int64

func (e ErrLowBackoffValue) Error() string { return fmt.Sprintf("back off should be at least 5: %d", e) }

// ErrParseTimeOut is for when the timeout cannot be parsed
type ErrParseTimeOut struct {
	Timeout string
	Err     error
}

func (e ErrParseTimeOut) Error() string {
	return fmt.Sprintf("parse timeout (%s): %s", e.Timeout, e.Err)
}

// ErrEndpointNotAvailable is the error when the endpoint is not available.
type ErrEndpointNotAvailable struct {
	Endpoint string
	Err      error
}

func (e ErrEndpointNotAvailable) Error() string {
	return fmt.Sprintf("endpoint (%s) not available: %s", e.Endpoint, e.Err)
}

// ErrInvalidIndexName is the error when the index name is invalid.
type ErrInvalidIndexName string

func (e ErrInvalidIndexName) Error() string {
	return fmt.Sprintf("Index name (%s) is not valid", string(e))
}

// ErrLowTimeout is the error when the interval is zero
type ErrLowTimeout time.Duration

func (e ErrLowTimeout) Error() string {
	return fmt.Sprintf("timeout should be more than 1 second: %d", e)
}
