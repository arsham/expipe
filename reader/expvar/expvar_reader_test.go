// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvar

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
)

func TestExpvarReaderErrors(t *testing.T) {
	log := lib.DiscardLogger()
	ctxReader := reader.NewMockCtxReader("nowhere")
	ctxReader.ContextReadFunc = func(ctx context.Context) (*http.Response, error) {
		return nil, fmt.Errorf("Error")
	}
	rdr, _ := NewExpvarReader(log, ctxReader)
	rdr.Start()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rdr.JobChan() <- ctx
	select {
	case res := <-rdr.ResultChan():
		if res.Res != nil {
			t.Errorf("expecting no results, got(%v)", res.Res)
		}
		if res.Err == nil {
			t.Error("expecting error, got nothing")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expecting an error result back, got nothing")
	}
}

func TestExpvarReaderReads(t *testing.T) {
	log := lib.DiscardLogger()
	testCase := `{"hey": 666}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, testCase)
	}))
	ctxReader := reader.NewMockCtxReader(ts.URL)
	rdr, _ := NewExpvarReader(log, ctxReader)
	rdr.Start()
	ctx, cancel := context.WithCancel(context.Background())
	rdr.JobChan() <- ctx
	res := <-rdr.ResultChan()
	if res.Err != nil {
		t.Errorf("Expecting no errors, got (%s)", res.Err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Res)
	if buf.String() != testCase {
		t.Errorf("want (%s), got (%s)", testCase, buf.String())
	}

	cancel()
}

func TestExpvarReaderClosesStream(t *testing.T) {
	log := lib.DiscardLogger()
	ctxReader := reader.NewMockCtxReader("nowhere")
	rdr, _ := NewExpvarReader(log, ctxReader)
	done := rdr.Start()
	ctx, cancel := context.WithCancel(context.Background())
	rdr.JobChan() <- ctx

	select {
	case <-rdr.ResultChan():
	default:
		close(rdr.JobChan())
	}
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("The channel was not closed in time")
	}
	cancel()
}
