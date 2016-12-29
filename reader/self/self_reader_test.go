// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package self_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/reader/self"
)

func setup(jobChanBuff, errorChanBuff int, message string) (red *self.Reader, errorChan chan communication.ErrorMessage, teardown func()) {
	log := lib.DiscardLogger()
	jobChan := make(chan context.Context, jobChanBuff)
	resultChan := make(chan *reader.ReadJobResult)
	errorChan = make(chan communication.ErrorMessage, errorChanBuff)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, message)
	}))
	red, _ = self.NewSelfReader(log, ts.URL, datatype.DefaultMapper(), jobChan, resultChan, errorChan, "test_self", "n/a", time.Hour, time.Hour)
	return red, errorChan, func() {
		ts.Close()
	}
}

func setupWithURL(url string, jobChanBuff, errorChanBuff int, message string) (red *self.Reader, errorChan chan communication.ErrorMessage) {
	log := lib.DiscardLogger()
	jobChan := make(chan context.Context, jobChanBuff)
	resultChan := make(chan *reader.ReadJobResult)
	errorChan = make(chan communication.ErrorMessage, errorChanBuff)
	red, _ = self.NewSelfReader(log, url, &datatype.MapConvertMock{}, jobChan, resultChan, errorChan, "my_reader", "example_type", time.Hour, time.Hour)
	return red, errorChan
}

func TestSelfReader(t *testing.T) {

	reader.TestReaderEssentials(t, func(testCase int) (reader.DataReader, chan communication.ErrorMessage, string, func()) {
		testMessage := `{"the key": "is the value!"}`

		switch testCase {
		case reader.GenericReaderReceivesJobTestCase:
			red, errorChan, teardown := setup(0, 0, testMessage)
			return red, errorChan, testMessage, teardown

		case reader.ReaderSendsResultTestCase:
			testMessage := `{"the key": "is the value!"}`
			red, errorChan, teardown := setup(0, 0, testMessage)
			return red, errorChan, testMessage, teardown

		case reader.ReaderReadsOnBufferedChanTestCase:
			red, errorChan, teardown := setup(10, 10, testMessage)
			return red, errorChan, testMessage, teardown

		case reader.ReaderDrainsAfterClosingContextTestCase:
			red, errorChan, teardown := setup(10, 10, testMessage)
			return red, errorChan, testMessage, teardown

		case reader.ReaderClosesTestCase:
			red, errorChan, teardown := setup(0, 0, testMessage)
			return red, errorChan, testMessage, teardown

		case reader.ReaderClosesWithBufferedChansTestCase:
			red, errorChan, teardown := setup(10, 10, testMessage)
			return red, errorChan, testMessage, teardown

		case reader.ReaderWithNoValidURLErrorsTestCase:
			red, errorChan := setupWithURL("nowhere", 0, 0, "")
			return red, errorChan, testMessage, nil

		default:
			return nil, nil, "", nil
		}
	})
}

func TestSelfReaderConstruction(t *testing.T) {
	reader.TestReaderConstruction(t, func(name, endpoint, typeName string, jobChan chan context.Context, resultChan chan *reader.ReadJobResult, interval time.Duration, timeout time.Duration) (reader.DataReader, error) {
		log := lib.DiscardLogger()
		errorChan := make(chan communication.ErrorMessage)
		return self.NewSelfReader(log, endpoint, datatype.DefaultMapper(), jobChan, resultChan, errorChan, name, typeName, interval, timeout)
	})
}

func TestSelfReaderEndpointManeuvers(t *testing.T) {
	reader.TestReaderEndpointManeuvers(t, func(testCase int, endpoint string) (reader.DataReader, chan communication.ErrorMessage) {
		switch testCase {
		case reader.ReaderErrorsOnEndpointDisapearsTestCase:
			log := lib.DiscardLogger()
			jobChan := make(chan context.Context)
			resultChan := make(chan *reader.ReadJobResult)
			errorChan := make(chan communication.ErrorMessage)
			red, _ := self.NewSelfReader(log, endpoint, &datatype.MapConvertMock{}, jobChan, resultChan, errorChan, "my_reader", "example_type", time.Second, time.Second)
			return red, errorChan

		default:
			return nil, nil
		}
	})
}
