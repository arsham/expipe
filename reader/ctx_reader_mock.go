// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"context"
	"net/http"
)

// MockCtxReader is the mocked version of CtxReader
type MockCtxReader struct {
	ContextReadFunc func(ctx context.Context) (*http.Response, error)
	url             string
}

// NewMockCtxReader instantiates a MockCtxReader object and sets the url to be read from ContextReadFunc
func NewMockCtxReader(url string) *MockCtxReader {
	return &MockCtxReader{
		url: url,
		ContextReadFunc: func(ctx context.Context) (*http.Response, error) {
			return http.Get(url)
		},
	}
}

// Get calls ContextReadFunc with ctx
func (m *MockCtxReader) Get(ctx context.Context) (*http.Response, error) {
	// not checking on purpose
	return m.ContextReadFunc(ctx)
}
