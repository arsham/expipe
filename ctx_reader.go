// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
    "context"
    "net/http"
)

// ContextReader reads from the url with the context
// Don't confuse it with the targetReader, that is for reading metrics!
type ContextReader interface {
    ContextRead(ctx context.Context) (*http.Response, error)
}

// CtxReader implements ContextReader interface
type CtxReader struct {
    url string
}

// NewCtxReader ..
func NewCtxReader(url string) *CtxReader {
    return &CtxReader{url}
}

// ContextRead returns DefaultClient errors
func (c *CtxReader) ContextRead(ctx context.Context) (*http.Response, error) {
    req, err := http.NewRequest("GET", c.url, nil)
    if err != nil {
        // Although it seems allright, but just in case
        return nil, err
    }
    req = req.WithContext(ctx)
    return http.DefaultClient.Do(req)
}
