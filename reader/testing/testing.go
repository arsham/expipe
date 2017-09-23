// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package testing is a test suit for readers. They should provide
// an object that implements the Constructor interface then run:
//
//    import reader_test "github.com/arsham/expipe/reader/testing"
//
//    for name, fn := range reader_test.TestSuites() {
//        t.Run(name, func(t *testing.T) {
//            r, err := reader_test.New(reader.SetName("test"))
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

	"github.com/arsham/expipe/reader"
)

// Constructor is an interface for setting up an object for testing.
type Constructor interface {
	reader.Constructor

	// TestServer should return a ready to use test server
	TestServer() *httptest.Server

	// Object should return the instantiated object
	Object() (reader.DataReader, error)
}

// TestSuites returns a map of test name to the runner function.
func TestSuites() map[string]func(t *testing.T, cons Constructor) {
	return map[string]func(t *testing.T, cons Constructor){
		"ShouldNotChangeTheInput": func(t *testing.T, cons Constructor) {
			testShouldNotChangeTheInput(t, cons)
		},

		"NameCheck": func(t *testing.T, cons Constructor) {
			testNameCheck(t, cons)
		},

		"TypeNameCheck": func(t *testing.T, cons Constructor) {
			testTypeNameCheck(t, cons)
		},

		"BackoffCheck": func(t *testing.T, cons Constructor) {
			testBackoffCheck(t, cons)
		},

		"IntervalCheck": func(t *testing.T, cons Constructor) {
			testIntervalCheck(t, cons)
		},

		"EndpointCheck": func(t *testing.T, cons Constructor) {
			testEndpointCheck(t, cons)
		},

		"ReceivesJob": func(t *testing.T, cons Constructor) {
			testReaderReceivesJob(t, cons)
		},

		"ReturnsSameID": func(t *testing.T, cons Constructor) {
			testReaderReturnsSameID(t, cons)
		},

		"PingingEndpoint": func(t *testing.T, cons Constructor) {
			pingingEndpoint(t, cons)
		},

		"ErrorsOnEndpointDisapears": func(t *testing.T, cons Constructor) {
			testReaderErrorsOnEndpointDisapears(t, cons)
		},

		"BacksOffOnEndpointGone": func(t *testing.T, cons Constructor) {
			testReaderBacksOffOnEndpointGone(t, cons)
		},

		"ReadingReturnsErrorIfNotPingedYet": func(t *testing.T, cons Constructor) {
			testReadingReturnsErrorIfNotPingedYet(t, cons)
		},
	}
}
