// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package self_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/reader/self"
	reader_test "github.com/arsham/expvastic/reader/testing"
)

func setup(message string) (red *self.Reader, teardown func()) {
	log := lib.DiscardLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, message)
	}))
	red, _ = self.New(log, ts.URL, datatype.DefaultMapper(), "test_self", "n/a", time.Hour, time.Hour, 10)
	return red, func() {
		ts.Close()
	}
}

func TestReaderConstruction(t *testing.T) {
	reader_test.TestReaderConstruction(t, func(name, endpoint, typeName string, interval time.Duration, timeout time.Duration, backoff int) (reader.DataReader, error) {
		log := lib.DiscardLogger()
		return self.New(log, endpoint, datatype.DefaultMapper(), name, typeName, interval, timeout, backoff)
	})
}

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
			return self.New(log, endpoint, datatype.DefaultMapper(), "self_reader", "self_reader", 1*time.Second, 1*time.Second, 10)

		case reader_test.ReaderBacksOffOnEndpointGoneTestCase:
			log := lib.DiscardLogger()
			return self.New(log, endpoint, datatype.DefaultMapper(), "self_reader", "self_reader", 1*time.Second, 1*time.Second, 5)

		default:
			return nil, nil
		}
	})
}
