// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
    "context"
    "net/http"
)

// ContextReader reads from the url with the specified context
type ContextReader interface {
    // Get reads from the url and returns DefaultClient errors
    // This operation's deadline and cancellation depends on ctx
    // You should close the Body when you finished reading from it
    Get(ctx context.Context) (*http.Response, error)
}

// CtxReader implements ContextReader interface
type CtxReader struct {
    url string
}

// NewCtxReader requires a sanitised url
func NewCtxReader(url string) *CtxReader {
    return &CtxReader{url}
}

// Get uses GET verb for retreiving the data
// TODO: implement other verbs
func (c *CtxReader) Get(ctx context.Context) (*http.Response, error) {
    req, err := http.NewRequest("GET", c.url, nil)
    if err != nil {
        // Although it should be allright, but just in case
        return nil, err
    }
    req = req.WithContext(ctx)
    return http.DefaultClient.Do(req)
}
