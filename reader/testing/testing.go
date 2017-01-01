// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/reader"
)

const (
	// GenericReaderReceivesJobTestCase invokes TestGenericReaderReceivesJob test
	GenericReaderReceivesJobTestCase = iota

	// ReaderSendsResultTestCase invokes TestReaderSendsResult test
	ReaderSendsResultTestCase

	// ReaderReadsOnBufferedChanTestCase invokes TestReaderReadsOnBufferedChan test
	ReaderReadsOnBufferedChanTestCase

	// ReaderWithNoValidURLErrorsTestCase invokes TestReaderWithNoValidURLErrors test
	ReaderWithNoValidURLErrorsTestCase

	// ReaderErrorsOnEndpointDisapearsTestCase invokes TestReaderErrorsOnEndpointDisapears test
	ReaderErrorsOnEndpointDisapearsTestCase

	// ReaderBacksOffOnEndpointGoneTestCase invokes TestReaderBacksOffOnEndpointGone test
	ReaderBacksOffOnEndpointGoneTestCase
)

// This file contains generic tests for various readers. You need to pass in your reader
// as a ready to use object, with all the necessary mocks, and these set of tests will do all
// the tests for you.
// IMPORTANT: you need to write the edge cases if they are not covered in this section.

// TestReaderEssentials runs all essential tests
func TestReaderEssentials(t *testing.T, setup func(testCase int) (red reader.DataReader, testMessage string, teardown func())) {
	t.Run("TestGenericReaderReceivesJob", func(t *testing.T) {
		red, _, _ := setup(GenericReaderReceivesJobTestCase)
		if red == nil {
			t.Fatal("You should implement GenericReaderReceivesJobTestCase")
		}
		testGenericReaderReceivesJob(t, red)
	})
}

// TestReaderEndpointManeuvers runs all tests regarding the endpoint changing state.
func TestReaderEndpointManeuvers(t *testing.T, setup func(testCase int, endpoint string) (red reader.DataReader, err error)) {
	t.Run("TestReaderErrorsOnEndpointDisapears", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		red, err := setup(ReaderErrorsOnEndpointDisapearsTestCase, ts.URL)
		if red == nil {
			t.Fatal("You should implement ReaderErrorsOnEndpointDisapearsTestCase")
		}
		testReaderErrorsOnEndpointDisapears(t, ts, red, err)
	})

	t.Run("TestReaderBacksOffOnEndpointGone", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		red, err := setup(ReaderBacksOffOnEndpointGoneTestCase, ts.URL)
		if red == nil {
			t.Fatal("You should implement ReaderBacksOffOnEndpointGoneTestCase")
		}
		testReaderBacksOffOnEndpointGone(t, ts, red, err)
	})
}

type setupFunc func(
	name string,
	typeName string,
	endpoint string,
	interval time.Duration,
	timeout time.Duration,
	backoff int,
) (reader.DataReader, error)

// TestReaderConstruction runs all essential tests on object construction.
func TestReaderConstruction(t *testing.T, setup setupFunc) {
	name := "the name"
	typeName := "my type"
	endpoint := "http://127.0.0.1:9200"
	interval := time.Hour
	timeout := time.Hour
	backoff := 5

	testShowNotChangeTheInput(t, setup, name, typeName, endpoint, interval, timeout, backoff)
	testEndpointCheck(t, setup, name, typeName, endpoint, interval, timeout, backoff)

	// Name Check
	red, err := setup("", endpoint, typeName, interval, timeout, backoff)
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%#v)", red)
	}
	if err != reader.ErrEmptyName {
		t.Errorf("expected ErrEmptyName, got (%v)", err)
	}

	// TypeName Check
	red, err = setup(name, endpoint, "", interval, timeout, backoff)
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%#v)", red)
	}
	if err != reader.ErrEmptyTypeName {
		t.Errorf("expected ErrEmptyTypeName, got (%v)", err)
	}

	// Backoff check
	red, err = setup(name, endpoint, typeName, interval, timeout, 3)
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%#v)", red)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(interface {
		LowBackoffValue()
	}); !ok {
		t.Errorf("expected ErrLowBackoffValue, got (%v)", err)
	}
	if !strings.Contains(err.Error(), "3") {
		t.Errorf("expected 3 be mentioned, got (%v)", err)
	}
}

func testShowNotChangeTheInput(t *testing.T, setup setupFunc, name string, typeName string, endpoint string, interval time.Duration, timeout time.Duration, backoff int) {
	red, err := setup(name, endpoint, typeName, interval, timeout, backoff)
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}

	if red.Name() != name {
		t.Errorf("given name should not be changed: %v", red.Name())
	}
	if red.TypeName() != typeName {
		t.Errorf("given type name should not be changed: %v", red.TypeName())
	}
	if red.Interval() != interval {
		t.Errorf("given interval should not be changed: %v", red.Timeout())
	}
	if red.Timeout() != timeout {
		t.Errorf("given timeout should not be changed: %v", red.Timeout())
	}
}

func testEndpointCheck(t *testing.T, setup setupFunc, name string, typeName string, endpoint string, interval time.Duration, timeout time.Duration, backoff int) {

	red, err := setup(name, "", typeName, interval, timeout, backoff)
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
	}
	if err != reader.ErrEmptyEndpoint {
		t.Errorf("expected ErrEmptyEndpoint, got (%v)", err)
	}

	invalidEndpoint := "this is invalid"
	red, err = setup(name, invalidEndpoint, typeName, interval, timeout, backoff)
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
	}
	if _, ok := err.(interface {
		InvalidEndpoint()
	}); !ok {
		t.Fatalf("expected ErrInvalidEndpoint, got (%v)", err)
	}
	if !strings.Contains(err.Error(), invalidEndpoint) {
		t.Errorf("expected (%s) be in the error message, got (%v)", invalidEndpoint, err)
	}

	unavailableEndpoint := "http://nowhere.localhost.localhost"
	red, err = setup(name, unavailableEndpoint, typeName, interval, timeout, backoff)
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
	}
	if err == nil {
		t.Fatal("expected ErrEndpointNotAvailable, got nil")
	}
	if _, ok := err.(interface {
		EndpointNotAvailable()
	}); !ok {
		t.Errorf("expected ErrEndpointNotAvailable, got (%v)", err)
	}
	if !strings.Contains(err.Error(), unavailableEndpoint) {
		t.Errorf("expected (%s) be in the error message, got (%v)", unavailableEndpoint, err)
	}

}

// testGenericReaderReceivesJob is a test helper to test the reader can receive jobs
func testGenericReaderReceivesJob(t *testing.T, red reader.DataReader) {
	ctx := context.Background()
	done := make(chan struct{})
	go func() {
		red.Read(communication.NewReadJob(ctx))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected the reader to receive the job, but it blocked")
	}
}

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
	// We don't know the backoff amount set in the reader, so we try 100 times until it closes.
	backedOff := false
	job := communication.NewReadJob(ctx)
	for i := 0; i < 10; i++ {
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
