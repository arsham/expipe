// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package testing is a test suit for recorders. They should provide
// an object that implements the Constructor interface then run:
//
//    import recorder_test "github.com/arsham/expipe/recorder/testing"
//
//    for name, fn := range recorder_test.TestSuites() {
//        t.Run(name, func(t *testing.T) {
//            r, err := recorder_test.New(recorder.SetName("test"))
//            if err != nil {
//                panic(err)
//            }
//            fn(t, &Construct{r})
//        })
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

	"github.com/arsham/expipe/recorder"
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
func TestSuites() map[string]func(t *testing.T, cons Constructor) {
	return map[string]func(t *testing.T, cons Constructor){
		"Construction": func(t *testing.T, cons Constructor) {
			testShouldNotChangeTheInput(t, cons)
		},
		"NameCheck": func(t *testing.T, cons Constructor) {
			testNameCheck(t, cons)
		},
		"BackoffCheck": func(t *testing.T, cons Constructor) {
			testBackoffCheck(t, cons)
		},
		"EndpointCheck": func(t *testing.T, cons Constructor) {
			testEndpointCheck(t, cons)
		},
		"ReceivesPayload": func(t *testing.T, cons Constructor) {
			testRecorderReceivesPayload(t, cons)
		},

		"SendsResult": func(t *testing.T, cons Constructor) {
			testRecorderSendsResult(t, cons)
		},

		"ErrorsOnUnavailableESServer": func(t *testing.T, cons Constructor) {
			if testing.Short() {
				t.Skip("Skipping ErrorsOnUnavailableESServer in short mod,")
			}
			testRecorderErrorsOnUnavailableEndpoint(t, cons)
		},

		"BacksOffOnEndpointGone": func(t *testing.T, cons Constructor) {
			testRecorderBacksOffOnEndpointGone(t, cons)
		},

		"RecordingReturnsErrorIfNotPingedYet": func(t *testing.T, cons Constructor) {
			testRecordingReturnsErrorIfNotPingedYet(t, cons)
		},
	}
}
