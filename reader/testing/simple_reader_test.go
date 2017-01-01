// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	reader_test "github.com/arsham/expvastic/reader/testing"
)

func setup(message string) (red *reader_test.SimpleReader, teardown func()) {
	log := lib.DiscardLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, message)
	}))
	red, _ = reader_test.NewSimpleReader(log, ts.URL, "reader_example", "reader_example", time.Second, time.Second, 5)
	return red, func() { ts.Close() }
}

func setupWithURL(url string, message string) (red *reader_test.SimpleReader) {
	log := lib.DiscardLogger()
	red, _ = reader_test.NewSimpleReader(log, url, "my_reader", "example_type", time.Hour, time.Hour, 10)
	return red
}

func TestReaderConstruction(t *testing.T) {
	reader_test.TestReaderConstruction(t, func(name, endpoint, typeName string, interval time.Duration, timeout time.Duration, backoff int) (reader.DataReader, error) {
		log := lib.DiscardLogger()
		return reader_test.NewSimpleReader(log, endpoint, name, typeName, time.Hour, time.Hour, backoff)
	})
}

// The purpose of these tests is to make sure the simple reader, which is a mock,
// works perfect, so other tests can rely on it.
func TestReaderCommunication(t *testing.T) {
	reader_test.TestReaderCommunication(t, func(testCase int) (reader.DataReader, string, func()) {
		testMessage := `{"the key": "is the value!"}`

		switch testCase {
		case reader_test.ReaderReceivesJobTestCase:
			red, teardown := setup(testMessage)
			return red, testMessage, teardown

		case reader_test.ReaderReturnsSameIDTestCase:
			red, teardown := setup(testMessage)
			return red, testMessage, teardown

		default:
			return nil, "", nil
		}
	})
}

func TestReaderEndpointManeuvers(t *testing.T) {
	reader_test.TestReaderEndpointManeuvers(t, func(testCase int, endpoint string) (reader.DataReader, error) {
		switch testCase {
		case reader_test.ReaderErrorsOnEndpointDisapearsTestCase:
			log := lib.DiscardLogger()
			return reader_test.NewSimpleReader(log, endpoint, "my_reader", "example_type", 1*time.Second, 1*time.Second, 10)

		case reader_test.ReaderBacksOffOnEndpointGoneTestCase:
			log := lib.DiscardLogger()
			return reader_test.NewSimpleReader(log, endpoint, "my_reader", "example_type", 10*time.Millisecond, 10*time.Millisecond, 5)

		default:
			return nil, nil
		}
	})
}
