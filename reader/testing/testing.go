// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package testing is a test suit for readers. They should provide
// an object that implements the Constructor interface then run:
//
//    import rt "github.com/arsham/expipe/reader/testing"
//    ....
//    func TestMyThings(t *testing.T) {
//        rt.TestSuites(t, func() (rt.Constructor, func()) {
//            r, err := rt.New(reader.SetName("test"))
//            if err != nil {
//                panic(err)
//            }
//            c := &Construct{Reader: r}
//            return c, func() { /*clean up code*/ }
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
// TestServer should return a ready to use test server
// Object should return the instantiated object
type Constructor interface {
	reader.Constructor
	TestServer() *httptest.Server
	Object() (reader.DataReader, error)
}

// TestSuites returns a map of test name to the runner function.
func TestSuites(t *testing.T, setup func() (Constructor, func())) {
	t.Parallel()
	t.Run("ShouldNotChangeTheInput", func(t *testing.T) {
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
	t.Run("TypeNameCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		typeNameCheck(t, cons)
	})
	t.Run("BackoffCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		backoffCheck(t, cons)
	})
	t.Run("IntervalCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		intervalCheck(t, cons)
	})
	t.Run("EndpointCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		endpointCheck(t, cons)
	})
	t.Run("ReceivesJob", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		readerReceivesJob(t, cons)
	})
	t.Run("ReturnsSameID", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		readerReturnsSameID(t, cons)
	})
	t.Run("PingingEndpoint", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		pingingEndpoint(t, cons)
	})
	t.Run("ErrorsOnEndpointDisapears", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		readerErrorsOnEndpointDisapears(t, cons)
	})
	t.Run("BacksOffOnEndpointGone", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		readerBacksOffOnEndpointGone(t, cons)
	})
	t.Run("ReadingReturnsErrorIfNotPingedYet", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		readingReturnsErrorIfNotPingedYet(t, cons)
	})
}
