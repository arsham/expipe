// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
)

func setup(jobChanBuff, resultChanBuff, errorChanBuff int, message string) (red *reader.SimpleReader, errorChan chan communication.ErrorMessage, teardown func()) {
	log := lib.DiscardLogger()
	jobChan := make(chan context.Context, jobChanBuff)
	resultChan := make(chan *reader.ReadJobResult, resultChanBuff)
	errorChan = make(chan communication.ErrorMessage, errorChanBuff)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, message)
	}))
	red, _ = reader.NewSimpleReader(log, ts.URL, jobChan, resultChan, errorChan, "reader_example", "reader_example", time.Second, time.Second)
	return red, errorChan, func() { ts.Close() }
}

func setupWithURL(url string, jobChanBuff, errorChanBuff int, message string) (red *reader.SimpleReader, errorChan chan communication.ErrorMessage) {
	log := lib.DiscardLogger()
	jobChan := make(chan context.Context, jobChanBuff)
	resultChan := make(chan *reader.ReadJobResult)
	errorChan = make(chan communication.ErrorMessage, errorChanBuff)
	red, _ = reader.NewSimpleReader(log, url, jobChan, resultChan, errorChan, "my_reader", "example_type", time.Hour, time.Hour)
	return red, errorChan
}

// The purpose of these tests is to make sure the simple reader, which is a mock,
// works perfect, so other tests can rely on it.
func TestSimpleReader(t *testing.T) {
	reader.TestReaderEssentials(t, func(testCase int) (reader.DataReader, chan communication.ErrorMessage, string, func()) {
		testMessage := `{"the key": "is the value!"}`

		switch testCase {
		case reader.GenericReaderReceivesJobTestCase:
			red, errorChan, teardown := setup(0, 1, 0, testMessage)
			return red, errorChan, testMessage, teardown

		case reader.ReaderSendsResultTestCase:
			testMessage := `{"the key": "is the value!"}`
			red, errorChan, teardown := setup(0, 0, 0, testMessage)
			return red, errorChan, testMessage, teardown

		case reader.ReaderReadsOnBufferedChanTestCase:
			red, errorChan, teardown := setup(10, 0, 10, testMessage)
			return red, errorChan, testMessage, teardown

		case reader.ReaderDrainsAfterClosingContextTestCase:
			red, errorChan, teardown := setup(10, 0, 10, testMessage)
			return red, errorChan, testMessage, teardown

		case reader.ReaderClosesTestCase:
			red, errorChan, teardown := setup(0, 0, 0, testMessage)
			return red, errorChan, testMessage, teardown

		case reader.ReaderClosesWithBufferedChansTestCase:
			red, errorChan, teardown := setup(10, 0, 10, testMessage)
			return red, errorChan, testMessage, teardown

		case reader.ReaderWithNoValidURLErrorsTestCase:
			red, errorChan := setupWithURL("nowhere", 0, 0, "")
			return red, errorChan, testMessage, nil

		default:
			return nil, nil, "", nil
		}
	})
}

func TestSimpleReaderConstruction(t *testing.T) {
	reader.TestReaderConstruction(t, func(name, endpoint, typeName string, jobChan chan context.Context, resultChan chan *reader.ReadJobResult, interval time.Duration, timeout time.Duration) (reader.DataReader, error) {
		log := lib.DiscardLogger()
		errorChan := make(chan communication.ErrorMessage)
		return reader.NewSimpleReader(log, endpoint, jobChan, resultChan, errorChan, name, typeName, time.Hour, time.Hour)
	})
}

func TestSimpleReaderEndpointManeuvers(t *testing.T) {
	reader.TestReaderEndpointManeuvers(t, func(testCase int, endpoint string) (reader.DataReader, chan communication.ErrorMessage) {
		switch testCase {
		case reader.ReaderErrorsOnEndpointDisapearsTestCase:
			log := lib.DiscardLogger()
			jobChan := make(chan context.Context)
			resultChan := make(chan *reader.ReadJobResult)
			errorChan := make(chan communication.ErrorMessage)
			red, _ := reader.NewSimpleReader(log, endpoint, jobChan, resultChan, errorChan, "my_reader", "example_type", 1*time.Second, 1*time.Second)
			return red, errorChan

		default:
			return nil, nil
		}
	})
}
