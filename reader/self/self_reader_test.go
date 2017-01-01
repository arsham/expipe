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
	red, _ = self.NewSelfReader(log, ts.URL, datatype.DefaultMapper(), "test_self", "n/a", time.Hour, time.Hour, 10)
	return red, func() {
		ts.Close()
	}
}

func TestSelfReader(t *testing.T) {

	reader_test.TestReaderEssentials(t, func(testCase int) (reader.DataReader, string, func()) {
		testMessage := `{"the key": "is the value!"}`

		switch testCase {
		case reader_test.GenericReaderReceivesJobTestCase:
			red, teardown := setup(testMessage)
			return red, testMessage, teardown

		case reader_test.ReaderSendsResultTestCase:
			testMessage := `{"the key": "is the value!"}`
			red, teardown := setup(testMessage)
			return red, testMessage, teardown

		case reader_test.ReaderReadsOnBufferedChanTestCase:
			red, teardown := setup(testMessage)
			return red, testMessage, teardown

		case reader_test.ReaderWithNoValidURLErrorsTestCase:
			log := lib.DiscardLogger()
			red, _ := self.NewSelfReader(log, "nowhere", &datatype.MapConvertMock{}, "my_reader", "example_type", time.Hour, time.Hour, 10)
			return red, testMessage, nil

		default:
			return nil, "", nil
		}
	})
}

func TestSelfReaderConstruction(t *testing.T) {
	reader_test.TestReaderConstruction(t, func(name, endpoint, typeName string, interval time.Duration, timeout time.Duration, backoff int) (reader.DataReader, error) {
		log := lib.DiscardLogger()
		return self.NewSelfReader(log, endpoint, datatype.DefaultMapper(), name, typeName, interval, timeout, backoff)
	})
}
