// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/recorder"
)

// The purpose of these tests is to make sure the simple recorder, which is a mock,
// works perfect, so other tests can rely on it.

func setupWithURL(URL string, errorChan chan communication.ErrorMessage) (ctx context.Context, rec *recorder.SimpleRecorder) {
	log := lib.DiscardLogger()
	ctx = context.Background()
	payloadChan := make(chan *recorder.RecordJob)
	rec, _ = recorder.NewSimpleRecorder(ctx, log, payloadChan, errorChan, "recorder_example", URL, "intexName", 10*time.Millisecond)
	return ctx, rec
}

func setup(errorChan chan communication.ErrorMessage) (ctx context.Context, rec *recorder.SimpleRecorder, teardown func()) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ctx, rec = setupWithURL(ts.URL, errorChan)
	return ctx, rec, func() { ts.Close() }
}

func TestSimpleRecorder(t *testing.T) {

	recorder.TestRecorderEssentials(t, func(testCase int) (context.Context, recorder.DataRecorder, error, chan communication.ErrorMessage, func()) {
		switch testCase {
		case recorder.RecorderReceivesPayloadTestCase:
			errorChan := make(chan communication.ErrorMessage)
			ctx, rec, teardown := setup(make(chan communication.ErrorMessage))
			return ctx, rec, nil, errorChan, teardown

		case recorder.RecorderSendsResultTestCase:
			errorChan := make(chan communication.ErrorMessage)
			ctx, rec, teardown := setup(make(chan communication.ErrorMessage))
			return ctx, rec, nil, errorChan, teardown

		case recorder.RecorderErrorsOnUnavailableEndpointTestCase:
			err := recorder.ErrEndpointNotAvailable{Endpoint: "nowhere", Err: nil} // this is a special case because it's a mock
			return context.TODO(), (*recorder.SimpleRecorder)(nil), err, nil, func() {}

		case recorder.RecorderClosesTestCase:
			errorChan := make(chan communication.ErrorMessage)
			ctx, rec, teardown := setup(make(chan communication.ErrorMessage))
			return ctx, rec, nil, errorChan, teardown

		default:
			return nil, nil, nil, nil, nil
		}
	})
}

func TestSimpleRecorderConstruction(t *testing.T) {
	recorder.TestRecorderConstruction(t, func(payloadChan chan *recorder.RecordJob, name, indexName string, timeout time.Duration) recorder.DataRecorder {
		log := lib.DiscardLogger()
		ctx := context.Background()
		rec, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, nil, name, "nowhere", indexName, timeout)
		return rec
	})
}
