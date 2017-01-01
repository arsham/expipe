// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/recorder"
)

const (
	// RecorderReceivesPayloadTestCase is for invoking TestRecorderReceivesPayload test
	RecorderReceivesPayloadTestCase = iota

	// RecorderSendsResultTestCase is for invoking TestRecorderSendsResult test
	RecorderSendsResultTestCase

	// RecorderClosesTestCase is for invoking TestRecorderCloses test
	RecorderClosesTestCase

	// RecorderErrorsOnUnavailableEndpointTestCase is for invoking TestRecorderErrorsOnUnavailableEndpoint test
	RecorderErrorsOnUnavailableEndpointTestCase

	// RecorderBacksOffOnEndpointGoneTestCase invokes TestRecorderBacksOffOnEndpointGone test
	RecorderBacksOffOnEndpointGoneTestCase

	// RecorderConstructionCasesTestCase is for invoking TestRecorderConstructionCases test
	RecorderConstructionCasesTestCase

	// RecorderErrorsOnInvalidEndpointTestCase is for invoking TestRecorderErrorsOnInvalidEndpoint test
	RecorderErrorsOnInvalidEndpointTestCase
)

// This file contains generic tests for various recorders. You need to pass in your recorder
// as a ready to use object, with all the necessary mocks, and these set of tests will do all
// the tests for you.
// Note 1: you need to write the edge cases if they are not covered in this section.
// Note 2: the recorder should not be started.

// TestRecorderEssentials runs all essential tests.
// The only case the error is needed is to check the endpoint on start up.
func TestRecorderEssentials(t *testing.T, setup func(testCase int) (ctx context.Context, rec recorder.DataRecorder, err error, teardown func())) {
	t.Run("TestRecorderReceivesPayload", func(t *testing.T) {
		ctx, rec, _, teardown := setup(RecorderReceivesPayloadTestCase)
		if rec == nil {
			t.Fatal("You should implement RecorderReceivesPayloadTestCase")
		}
		defer teardown()
		testRecorderReceivesPayload(ctx, t, rec)
	})

	t.Run("TestRecorderSendsResult", func(t *testing.T) {
		ctx, rec, _, teardown := setup(RecorderSendsResultTestCase)
		if rec == nil {
			t.Fatal("You should implement RecorderSendsResultTestCase")
		}
		defer teardown()
		testRecorderSendsResult(ctx, t, rec)
	})
}

// TestRecorderEndpointManeuvers runs all tests regarding the endpoint changing state.
func TestRecorderEndpointManeuvers(t *testing.T, setup func(testCase int) (ctx context.Context, rec recorder.DataRecorder, err error, teardown func())) {
	t.Run("TestRecorderErrorsOnUnavailableESServer", func(t *testing.T) {
		_, rec, err, _ := setup(RecorderErrorsOnUnavailableEndpointTestCase)
		if rec == nil {
			t.Fatal("You should implement RecorderErrorsOnUnavailableEndpointTestCase")
		}
		testRecorderErrorsOnUnavailableEndpoint(t, rec, err)
	})

	t.Run("TestRecorderBacksOffOnEndpointGone", func(t *testing.T) {
		_, rec, _, teardown := setup(RecorderBacksOffOnEndpointGoneTestCase)
		if rec == nil {
			t.Fatal("You should implement RecorderBacksOffOnEndpointGoneTestCase")
		}
		testRecorderBacksOffOnEndpointGone(t, rec, teardown)
	})
}

func isTravis() bool {
	return os.Getenv("TRAVIS") != ""
}

type setupFunc func(
	testCase int,
	name,
	endpoint,
	indexName string,
	timeout time.Duration,
	backoff int,
) (recorder.DataRecorder, error)

// TestRecorderConstruction runs all essential tests on object construction.
func TestRecorderConstruction(t *testing.T, setup setupFunc) {
	t.Run("TestRecorderConstructionCases", func(t *testing.T) {
		testRecorderConstructionCases(t, setup)
	})

	t.Run("TestRecorderErrorsOnInvalidEndpoint", func(t *testing.T) {
		testRecorderErrorsOnInvalidEndpoint(t, setup)
	})
}

func testRecorderConstructionCases(t *testing.T, setup setupFunc) {
	name := "the name"
	indexName := "index_name"
	endpoint := "http://127.0.0.1:9200"
	backoff := 5

	timeout := 10 * time.Millisecond
	if isTravis() {
		timeout = 10 * time.Second
	}
	rec, _ := setup(RecorderConstructionCasesTestCase, name, endpoint, indexName, timeout, backoff)

	if rec.Name() != name {
		t.Errorf("given name should not be changed: %v", rec.Name())
	}
	if rec.IndexName() != indexName {
		t.Errorf("given index name should not be changed: %v", rec.IndexName())
	}
	if rec.Timeout() != timeout {
		t.Errorf("given timeout should not be changed: %v", rec.Timeout())
	}

	// Backoff check
	rec, err := setup(RecorderConstructionCasesTestCase, name, endpoint, indexName, timeout, 3)
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("expected nil, got (%#v)", rec)
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

func testRecorderErrorsOnInvalidEndpoint(t *testing.T, setup setupFunc) {
	name := "the name"
	indexName := "index_name"
	backoff := 5

	timeout := 10 * time.Millisecond
	if isTravis() {
		timeout = 10 * time.Second
	}
	invalidEndpoint := "this is invalid"
	rec, err := setup(RecorderErrorsOnInvalidEndpointTestCase, name, invalidEndpoint, indexName, timeout, backoff)
	if rec == nil {
		t.Fatal("You should implement RecorderErrorsOnInvalidEndpointTestCase")
	}
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("expected nil, got (%v)", rec)
	}
	if _, ok := err.(interface {
		InvalidEndpoint()
	}); !ok {
		t.Fatalf("expected ErrInvalidEndpoint, got (%v)", err)
	}
	if !strings.Contains(err.Error(), invalidEndpoint) {
		t.Errorf("expected (%s) be in the error message, got (%v)", invalidEndpoint, err)
	}
}

// testRecorderReceivesPayload tests the recorder receives the payload correctly.
func testRecorderReceivesPayload(ctx context.Context, t *testing.T, rec recorder.DataRecorder) {
	p := datatype.NewContainer([]datatype.DataType{})
	payload := &recorder.RecordJob{
		ID:        communication.NewJobID(),
		Payload:   p,
		IndexName: "my index",
		TypeName:  "my type",
		Time:      time.Now(),
	}

	received := make(chan struct{})
	go func() {
		rec.Record(ctx, payload)
		received <- struct{}{}
	}()

	select {
	case <-received:
	case <-time.After(5 * time.Second):
		t.Error("expected the recorder to receive the payload, but it blocked")
	}
}

// testRecorderSendsResult tests the recorder send the results to the endpoint.
func testRecorderSendsResult(ctx context.Context, t *testing.T, rec recorder.DataRecorder) {
	p := datatype.NewContainer([]datatype.DataType{&datatype.StringType{Key: "test", Value: "test"}})
	payload := &recorder.RecordJob{
		ID:        communication.NewJobID(),
		Payload:   p,
		IndexName: "my index",
		TypeName:  "my type",
		Time:      time.Now(),
	}

	err := rec.Record(ctx, payload)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}

// testRecorderErrorsOnUnavailableEndpoint tests the recorder errors for bad URL.
func testRecorderErrorsOnUnavailableEndpoint(t *testing.T, rec recorder.DataRecorder, err error) {
	if err == nil {
		t.Error("want error, got nil")
	}
	if _, ok := err.(interface {
		EndpointNotAvailable()
	}); !ok {
		t.Errorf("want EndpointNotAvailable, got (%v)", err)
	}
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("want (nil), got (%v)", rec)
	}
}

// testRecorderBacksOffOnEndpointGone is a helper to test the recorder backs off when the endpoint goes away.
func testRecorderBacksOffOnEndpointGone(t *testing.T, rec recorder.DataRecorder, teardown func()) {
	ctx := context.Background()
	teardown()
	p := datatype.NewContainer([]datatype.DataType{})
	payload := &recorder.RecordJob{
		ID:        communication.NewJobID(),
		Payload:   p,
		IndexName: "my index",
		TypeName:  "my type",
		Time:      time.Now(),
	}

	// We don't know the backoff amount set in the recorder, so we try 100 times until it closes.
	backedOff := false
	for i := 0; i < 100; i++ {
		err := rec.Record(ctx, payload)
		if err == recorder.ErrBackoffExceeded {
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
		rec.Record(ctx, payload)
		close(done)
	}()
	select {
	case <-done:
		// good one!
	case <-time.After(20 * time.Millisecond):
		t.Error("expected the recorder to be gone")
	}
}
