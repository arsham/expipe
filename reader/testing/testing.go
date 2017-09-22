// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package testing is a test suit for readers. They should provide
// an object that implements the Constructor interface then run:
//
//    import reader_test "github.com/arsham/expipe/reader/testing"
//
//    func TestExpvar(t *testing.T) {
//    	reader_testing.TestReader(t, &Construct{})
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

	"github.com/arsham/expipe/reader"
)

// Constructor is an interface for setting up an object for testing.
type Constructor interface {
	// SetName is for setting the Name
	SetName(string)

	// SetTypename is for setting the Typename
	SetTypename(string)

	// SetEndpoint is for setting the Endpoint
	SetEndpoint(string)

	// SetInterval is for setting the Interval
	SetInterval(time.Duration)

	// SetTimeout is for setting the Timeout
	SetTimeout(time.Duration)

	// SetBackoff is for setting the Backoff
	SetBackoff(int)

	// TestServer should return a ready to use test server
	TestServer() *httptest.Server

	// Object should return the instantiated object
	Object() (reader.DataReader, error)
}

// TestReader runs all essential tests on object construction.
func TestReader(t *testing.T, cons Constructor) {
	t.Run("ShowNotChangeTheInput", func(t *testing.T) {
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

	t.Run("ReceivesJob", func(t *testing.T) {
		testReaderReceivesJob(t, cons)
	})

	t.Run("ReturnsSameID", func(t *testing.T) {
		testReaderReturnsSameID(t, cons)
	})

	t.Run("PingingEndpoint", func(t *testing.T) {
		pingingEndpoint(t, cons)
	})

	t.Run("ErrorsOnEndpointDisapears", func(t *testing.T) {
		testReaderErrorsOnEndpointDisapears(t, cons)
	})

	t.Run("BacksOffOnEndpointGone", func(t *testing.T) {
		testReaderBacksOffOnEndpointGone(t, cons)
	})

	t.Run("ReadingReturnsErrorIfNotPingedYet", func(t *testing.T) {
		testReadingReturnsErrorIfNotPingedYet(t, cons)
	})
}
