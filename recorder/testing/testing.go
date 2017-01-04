// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package testing is a test suit for recorders. They should provide
// an object that implements the Constructor interface then run:
//
//    import recorder_test "github.com/arsham/expvastic/recorder/testing"
//
//    func TestElasticsearch(t *testing.T) {
//    	recorder_testing.TestRecorder(t, &Construct{})
//    }
//
// The test suit will pick it up and does all the tests.
//
// Important Note
//
// You need to write the edge cases if they are not covered in this section.
//
package testing

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic/recorder"
)

// Constructor is an interface for setting up an object for testing.
type Constructor interface {
	// SetName is for setting the Name
	SetName(string)

	// SetIndexName is for setting the IndexName
	SetIndexName(string)

	// SetEndpoint is for setting the Endpoint
	SetEndpoint(string)

	// SetTimeout is for setting the Timeout
	SetTimeout(time.Duration)

	// SetBackoff is for setting the Backoff
	SetBackoff(int)

	// ValidEndpoints should return a list of valid endpoints
	ValidEndpoints() []string

	// InvalidEndpoints should return a list of invalid endpoints
	InvalidEndpoints() []string

	// TestServer should return a ready to use test server
	TestServer() *httptest.Server

	// Object should return the instantiated object
	Object() (recorder.DataRecorder, error)
}

// TestRecorder runs all essential tests on object construction.
func TestRecorder(t *testing.T, cons Constructor) {
	t.Run("Construction", func(t *testing.T) {
		testShowNotChangeTheInput(t, cons)
	})
	t.Run("NameCheck", func(t *testing.T) {
		testNameCheck(t, cons)
	})
	t.Run("BackoffCheck", func(t *testing.T) {
		testBackoffCheck(t, cons)
	})
	t.Run("EndpointCheck", func(t *testing.T) {
		testEndpointCheck(t, cons)
	})
	t.Run("ReceivesPayload", func(t *testing.T) {
		testRecorderReceivesPayload(t, cons)
	})

	t.Run("SendsResult", func(t *testing.T) {
		testRecorderSendsResult(t, cons)
	})

	t.Run("ErrorsOnUnavailableESServer", func(t *testing.T) {
		testRecorderErrorsOnUnavailableEndpoint(t, cons)
	})

	t.Run("BacksOffOnEndpointGone", func(t *testing.T) {
		testRecorderBacksOffOnEndpointGone(t, cons)
	})

	t.Run("RecordingReturnsErrorIfNotPingedYet", func(t *testing.T) {
		testRecordingReturnsErrorIfNotPingedYet(t, cons)
	})
}
