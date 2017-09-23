// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arsham/expipe/internal/token"
	"github.com/arsham/expipe/reader"
	"github.com/pkg/errors"
)

// pingingEndpoint is a helper to test the reader errors when the endpoint goes away.
func pingingEndpoint(t *testing.T, cons Constructor) {
	ts := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	ts.Close()
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetEndpoint(ts.URL)
	cons.SetInterval(time.Millisecond)
	cons.SetTimeout(time.Second)
	red, err := cons.Object()
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}

	if err := red.Ping(); err == nil {
		t.Error("expected error, got nil")
	}

	if err := red.Ping(); err == nil {
		t.Errorf("expected an error, got nil")
	} else if _, ok := errors.Cause(err).(reader.ErrEndpointNotAvailable); !ok {
		t.Errorf("want ErrEndpointNotAvailable, got (%v)", err)
	}

	unavailableEndpoint := "http://192.168.255.255"
	cons.SetEndpoint(unavailableEndpoint)
	red, _ = cons.Object()

	if err = red.Ping(); err == nil {
		t.Fatal("expected ErrEndpointNotAvailable, got nil")
	}
	err = errors.Cause(err)
	if _, ok := err.(reader.ErrEndpointNotAvailable); !ok {
		t.Errorf("expected ErrEndpointNotAvailable, got (%v)", err)
	}

	if !strings.Contains(err.Error(), unavailableEndpoint) {
		t.Errorf("expected (%s) be in the error message, got (%v)", unavailableEndpoint, err)
	}
}

// testReaderErrorsOnEndpointDisapears is a helper to test the reader errors when the endpoint goes away.
func testReaderErrorsOnEndpointDisapears(t *testing.T, cons Constructor) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetEndpoint(ts.URL)
	cons.SetInterval(time.Hour)
	cons.SetTimeout(time.Hour)
	cons.SetBackoff(5)
	red, err := cons.Object()
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}

	err = red.Ping()
	if err != nil {
		t.Fatal(err)
	}
	ts.Close()
	ctx := context.Background()
	done := make(chan struct{})
	go func() {
		result, err := red.Read(token.New(ctx))
		if err == nil {
			t.Error("want error, got nil")
			return
		}
		err = errors.Cause(err)
		if _, ok := err.(reader.ErrEndpointNotAvailable); !ok {
			t.Errorf("want ErrEndpointNotAvailable, got (%v)", err)
		}
		if !strings.Contains(err.Error(), ts.URL) {
			t.Errorf("want (%s) in error message, got (%s)", ts.URL, err)
		}
		if result != nil {
			t.Errorf("didn't expect to receive a data back, got (%v)", result)
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
func testReaderBacksOffOnEndpointGone(t *testing.T, cons Constructor) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetEndpoint(ts.URL)
	cons.SetInterval(time.Hour)
	cons.SetTimeout(time.Hour)
	cons.SetBackoff(5)
	red, err := cons.Object()
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}

	err = red.Ping()
	if err != nil {
		t.Fatal(err)
	}
	ts.Close()

	ctx := context.Background()
	backedOff := false
	job := token.New(ctx)
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

// testReadingReturnsErrorIfNotPingedYet is a helper to test the reader returns an error
// if the caller hasn't called the Ping() method.
func testReadingReturnsErrorIfNotPingedYet(t *testing.T, cons Constructor) {
	ctx := context.Background()
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetInterval(time.Second)
	cons.SetTimeout(time.Second)
	cons.SetBackoff(5)
	red, err := cons.Object()
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}

	job := token.New(ctx)

	res, err := red.Read(job)
	if err != reader.ErrPingNotCalled {
		t.Errorf("want ErrHasntCalledPing, got (%v)", err)
	}
	if res != nil {
		t.Errorf("want an empty result, got (%v)", err)
	}
}
