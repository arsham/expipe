// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import "fmt"

// TODO: apply these errors

var (
	// ErrEmptyName is the error when the package name is empty.
	ErrEmptyName = fmt.Errorf("name cannot be empty")

	// ErrEmptyEndpoint is the error when the given endpoint is empty.
	ErrEmptyEndpoint = fmt.Errorf("endpoint cannot be empty")

	// ErrEmptyTypeName is the error when the type_name is an empty string.
	ErrEmptyTypeName = fmt.Errorf("type_name cannot be empty")
)

// ErrInvalidEndpoint is the error when the endpoint is not a valid url
type ErrInvalidEndpoint string

// InvalidEndpoint defines the behaviour of the error
func (ErrInvalidEndpoint) InvalidEndpoint() {}
func (e ErrInvalidEndpoint) Error() string  { return fmt.Sprintf("invalid endpoint: %s", string(e)) }

// ErrEndpointNotAvailable is the error when the endpoint is not available.
type ErrEndpointNotAvailable struct {
	Endpoint string
	Err      error
}

// EndpointNotAvailable defines the behaviour of the error
func (ErrEndpointNotAvailable) EndpointNotAvailable() {}
func (e ErrEndpointNotAvailable) Error() string {
	return fmt.Sprintf("endpoint (%s) not available: %s", e.Endpoint, e.Err)
}
