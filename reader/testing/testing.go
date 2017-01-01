// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic/reader"
)

const (
	// ReaderReceivesJobTestCase invokes TestReaderReceivesJob test
	ReaderReceivesJobTestCase = iota

	// ReaderReturnsSameIDTestCase invokes TestReaderReturnsSameID test
	ReaderReturnsSameIDTestCase

	// ReaderErrorsOnEndpointDisapearsTestCase invokes TestReaderErrorsOnEndpointDisapears test
	ReaderErrorsOnEndpointDisapearsTestCase

	// ReaderBacksOffOnEndpointGoneTestCase invokes TestReaderBacksOffOnEndpointGone test
	ReaderBacksOffOnEndpointGoneTestCase
)

// This file contains generic tests for various readers. You need to pass in your reader
// as a ready to use object, with all the necessary mocks, and these set of tests will do all
// the tests for you.
// IMPORTANT: you need to write the edge cases if they are not covered in this section.

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
	testNameCheck(t, setup, name, typeName, endpoint, interval, timeout, backoff)
	testBackoffCheck(t, setup, name, typeName, endpoint, interval, timeout, backoff)
}

// TestReaderCommunication runs all essential tests
func TestReaderCommunication(t *testing.T, setup func(testCase int) (red reader.DataReader, testMessage string, teardown func())) {
	t.Run("TestReaderReceivesJob", func(t *testing.T) {
		red, _, _ := setup(ReaderReceivesJobTestCase)
		if red == nil {
			t.Fatal("You should implement ReaderReceivesJobTestCase")
		}
		testReaderReceivesJob(t, red)
	})

	t.Run("TestReaderReturnsSameID", func(t *testing.T) {
		red, _, _ := setup(ReaderReturnsSameIDTestCase)
		if red == nil {
			t.Fatal("You should implement ReaderReturnsSameIDTestCase")
		}
		testReaderReturnsSameID(t, red)
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
