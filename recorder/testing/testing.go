// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package testing is a test suit for recorders. They should provide
// an object that implements the Constructor interface then run:
//
//    import recorder_test "github.com/arsham/expipe/recorder/testing"
//    ....
//    r, err := recorder_test.New()
//    if err != nil {
//        panic(err)
//    }
//    c := &Construct{r, getTestServer()}
//    recorder_test.TestSuites(t, c)
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

	"github.com/arsham/expipe/recorder"
	gin "github.com/onsi/ginkgo"
)

// Constructor is an interface for setting up an object for testing.
type Constructor interface {
	recorder.Constructor

	// ValidEndpoints should return a list of valid endpoints
	ValidEndpoints() []string

	// InvalidEndpoints should return a list of invalid endpoints
	InvalidEndpoints() []string

	// TestServer should return a ready to use test server
	TestServer() *httptest.Server

	// Object should return the instantiated object
	Object() (recorder.DataRecorder, error)
}

// TestSuites returns a map of test name to the runner function.
func TestSuites(t *testing.T, cons Constructor) {

	t.Run("Construction", func(*testing.T) {
		gin.Describe("Checking input", func() {
			testShouldNotChangeTheInput(cons)
		})
	})
	t.Run("NameCheck", func(*testing.T) {
		gin.Describe("Checking name", func() {
			testNameCheck(cons)
		})
	})
	t.Run("IndexNameCheck", func(*testing.T) {
		gin.Describe("Checking index name", func() {
			testIndexNameCheck(cons)
		})
	})
	t.Run("BackoffCheck", func(*testing.T) {
		gin.Describe("Checking backoff value", func() {
			testBackoffCheck(cons)
		})
	})
	t.Run("EndpointCheck", func(*testing.T) {
		gin.Describe("Checking endpoint value", func() {
			testEndpointCheck(cons)
		})
	})
	t.Run("ReceivesPayload", func(*testing.T) {
		gin.Describe("Sending payload to recorder", func() {
			testRecorderReceivesPayload(cons)
		})
	})

	t.Run("SendsResult", func(*testing.T) {
		gin.Describe("Sending results", func() {
			testRecorderSendsResult(cons)
		})
	})

	t.Run("ErrorsOnUnavailableESServer", func(*testing.T) {
		gin.Describe("Errors", func() {
			testRecorderErrorsOnUnavailableEndpoint(cons)
		})
	})

	t.Run("BacksOffOnEndpointGone", func(*testing.T) {
		if testing.Short() {
			t.Skip("Skipping BacksOffOnEndpointGone in short mod,")
		}
		gin.Describe("Backing off when the endpoint is gone", func() {
			testRecorderBacksOffOnEndpointGone(cons)
		})
	})

	t.Run("RecordingReturnsErrorIfNotPingedYet", func(*testing.T) {
		gin.Describe("Recording without pinging", func() {
			testRecordingReturnsErrorIfNotPingedYet(cons)
		})
	})
}
