// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package testing is a test suit for recorders. They should provide
// an object that implements the Constructor interface then run:
//
//    import rt "github.com/arsham/expipe/recorder/testing"
//    ....
//    r, err := rt.New()
//    if err != nil {
//        panic(err)
//    }
//    c := &Construct{r, getTestServer()}
//    rt.TestSuites(t, c, func() {/*clean up code*/})
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
)

// Constructor is an interface for setting up an object for testing.
// TestServer() should return a ready to use test server
type Constructor interface {
	recorder.Constructor
	ValidEndpoints() []string
	InvalidEndpoints() []string
	TestServer() *httptest.Server
	Object() (recorder.DataRecorder, error)
}

// TestSuites returns a map of test name to the runner function.
func TestSuites(t *testing.T, setup func() (Constructor, func())) {
	t.Parallel()
	t.Run("Construction", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		shouldNotChangeTheInput(t, cons)
	})
	t.Run("NameCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		nameCheck(t, cons)
	})
	t.Run("IndexNameCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		indexNameCheck(t, cons)
	})
	t.Run("BackoffCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		backoffCheck(t, cons)
	})
	t.Run("EndpointCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		endpointCheck(t, cons)
	})
	t.Run("ReceivesPayload", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		recorderReceivesPayload(t, cons)
	})
	t.Run("SendsResult", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		recorderSendsResult(t, cons)
	})
	t.Run("ErrorsOnUnavailableESServer", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		recorderErrorsOnUnavailableEndpoint(t, cons)
	})
	t.Run("BacksOffOnEndpointGone", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		recorderBacksOffOnEndpointGone(t, cons)
	})
	t.Run("RecordingReturnsErrorIfNotPingedYet", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		recordingReturnsErrorIfNotPingedYet(t, cons)
	})
}
