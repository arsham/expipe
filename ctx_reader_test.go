// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
    "bytes"
    "context"
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/arsham/expvastic"
)

type MockCtxReader struct {
    ContextReadFunc func(ctx context.Context) (*http.Response, error)
    url             string
}

func NewMockCtxReader(url string) *MockCtxReader {
    return &MockCtxReader{
        url: url,
        ContextReadFunc: func(ctx context.Context) (*http.Response, error) {
            return http.Get(url)
        },
    }
}

func (m *MockCtxReader) ContextRead(ctx context.Context) (*http.Response, error) {
    // not checking on purpose
    return m.ContextReadFunc(ctx)
}

func TestContextReaderErrors(t *testing.T) {
    resp := "my response"
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, resp)
    }))
    ctxReader := expvastic.NewCtxReader("bad url")
    ctx, cancel := context.WithCancel(context.Background())
    res, err := ctxReader.ContextRead(ctx)
    if err == nil {
        t.Error("expected error, got nothing")
    }
    if res != nil {
        t.Errorf("expected empty response, got (%v)", res)
    }

    ctxReader = expvastic.NewCtxReader(ts.URL)
    res, err = ctxReader.ContextRead(ctx)
    if err != nil {
        t.Errorf("expected no errors, got (%s)", err)
    }
    if res == nil {
        t.Errorf("expected (%s), got nil", res)
    }
    defer res.Body.Close()
    buf := new(bytes.Buffer)
    buf.ReadFrom(res.Body)
    if buf.String() != resp {
        t.Errorf("expected (%s), got (%s)", resp, buf.String())
    }
    cancel()
}
