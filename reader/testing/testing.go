// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package testing is a test suit for readers. They should provide
// an object that implements the Constructor interface then run:
//
//    import reader_test "github.com/arsham/expipe/reader/testing"
//    ....
//    r, err := reader_test.New(reader.SetName("test"))
//    if err != nil {
//        panic(err)
//    }
//    c := &Construct{r}
//    reader_test.TestSuites(t, c)
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
	. "github.com/onsi/ginkgo"
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
func TestSuites(t *testing.T, cons Constructor) {

	t.Run("Should NotChangeTheInput", func(t *testing.T) {
		Describe("Checking input", func() {
			testShouldNotChangeTheInput(cons)
		})
	})

	t.Run("NameCheck", func(t *testing.T) {
		Describe("Checking name and index name", func() {
			testNameCheck(cons)
		})
	})

	t.Run("TypeNameCheck", func(t *testing.T) {
		testTypeNameCheck(cons)
	})

	t.Run("BackoffCheck", func(t *testing.T) {
		Describe("Checking backoff value", func() {
			testBackoffCheck(cons)
		})
	})

	t.Run("IntervalCheck", func(t *testing.T) {
		Describe("Checking interval value", func() {
			testIntervalCheck(cons)
		})
	})

	t.Run("EndpointCheck", func(t *testing.T) {
		Describe("Checking endpoint value", func() {
			testEndpointCheck(cons)
		})
	})

	t.Run("ReceivesJob", func(t *testing.T) {
		Describe("Receiving payload", func() {
			testReaderReceivesJob(cons)
		})
	})

	t.Run("ReturnsSameID", func(t *testing.T) {
		Describe("Returning the same job ID", func() {
			testReaderReturnsSameID(cons)
		})
	})

	t.Run("PingingEndpoint", func(t *testing.T) {
		Describe("Pinging the endpoint", func() {
			pingingEndpoint(cons)
		})
	})

	t.Run("ErrorsOnEndpointDisapears", func(t *testing.T) {
		Describe("Backing off when the endpoint disappears", func() {
			testReaderErrorsOnEndpointDisapears(cons)
		})
	})

	t.Run("BacksOffOnEndpointGone", func(t *testing.T) {
		Describe("Backing off when the endpoint is gone", func() {
			testReaderBacksOffOnEndpointGone(cons)
		})
	})

	t.Run("ReadingReturnsErrorIfNotPingedYet", func(t *testing.T) {
		Describe("Reading without pinging", func() {
			testReadingReturnsErrorIfNotPingedYet(cons)
		})
	})
}
