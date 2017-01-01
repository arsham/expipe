// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/reader"
)

// testReaderErrorsOnEndpointDisapears is a helper to test the reader errors when the endpoint goes away.
func testReaderErrorsOnEndpointDisapears(t *testing.T, ts *httptest.Server, red reader.DataReader, err error) {
	var res *reader.ReadJobResult
	ctx := context.Background()
	done := make(chan struct{})
	ts.Close()
	go func() {
		result, err := red.Read(communication.NewReadJob(ctx))
		if err == nil {
			t.Errorf("want error, got (%s)", err)
		}
		if result != nil {
			t.Errorf("didn't expect to receive a data back, got (%v)", res)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(20 * time.Millisecond):
		t.Error("expected to receive an error, nothing received")
	}
}

// testReaderBacksOffOnEndpointGone is a helper to test the reader backs off when the endpoint goes away.
func testReaderBacksOffOnEndpointGone(t *testing.T, ts *httptest.Server, red reader.DataReader, err error) {
	ctx := context.Background()
	ts.Close()
	backedOff := false
	job := communication.NewReadJob(ctx)
	// We don't know the backoff amount set in the reader, so we try 100 times until it closes.
	for i := 0; i < 100; i++ {
		_, err := red.Read(job)
		if err == reader.ErrBackoffExceeded {
			backedOff = true
			break
		}
	}
	if !backedOff {
		t.Error("expected to receive a ErrBackoffExceeded")
	}

	// sending another job, it should block
	done := make(chan struct{})
	go func() {
		red.Read(job)
		close(done)
	}()
	select {
	case <-done:
		// good one!
	case <-time.After(20 * time.Millisecond):
		t.Error("expected the recorder to be gone")
	}
}
