// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/recorder"
	recorder_testing "github.com/arsham/expvastic/recorder/testing"
)

// The purpose of these tests is to make sure the simple recorder, which is a mock,
// works perfect, so other tests can rely on it.

func setupWithURL(URL string, backoff int) (ctx context.Context, rec *recorder_testing.SimpleRecorder, err error) {
	log := lib.DiscardLogger()
	ctx = context.Background()
	rec, err = recorder_testing.NewSimpleRecorder(ctx, log, "recorder_example", URL, "intexName", 10*time.Millisecond, backoff)
	return ctx, rec, err
}

func setup(backoff int) (ctx context.Context, rec *recorder_testing.SimpleRecorder, err error, teardown func()) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ctx, rec, err = setupWithURL(ts.URL, backoff)
	return ctx, rec, err, func() { ts.Close() }
}

func TestRecorderCommunication(t *testing.T) {
	recorder_testing.TestRecorderCommunication(t, func(testCase int) (context.Context, recorder.DataRecorder, error, func()) {
		switch testCase {
		case recorder_testing.RecorderReceivesPayloadTestCase:
			ctx, rec, err, teardown := setup(5)
			return ctx, rec, err, teardown

		case recorder_testing.RecorderSendsResultTestCase:
			ctx, rec, err, teardown := setup(5)
			return ctx, rec, err, teardown

		case recorder_testing.RecorderClosesTestCase:
			ctx, rec, err, teardown := setup(5)
			return ctx, rec, err, teardown

		default:
			return nil, nil, nil, nil
		}
	})
}

func TestSimpleRecorderConstruction(t *testing.T) {
	recorder_testing.TestRecorderConstruction(t, func(testCase int, name, endpoint, indexName string, timeout time.Duration, backoff int) (recorder.DataRecorder, error) {
		switch testCase {
		case recorder_testing.RecorderConstructionCasesTestCase:
			log := lib.DiscardLogger()
			ctx := context.Background()
			return recorder_testing.NewSimpleRecorder(ctx, log, name, endpoint, indexName, timeout, backoff)

		case recorder_testing.RecorderErrorsOnInvalidEndpointTestCase:
			return (*recorder_testing.SimpleRecorder)(nil), recorder.ErrInvalidEndpoint(endpoint) // this is a special case because it's a mock

		default:
			return nil, nil
		}
	})
}

func TestSimpleRecorderEndpointManeuvers(t *testing.T) {
	t.Parallel()
	recorder_testing.TestRecorderEndpointManeuvers(t, func(testCase int) (context.Context, recorder.DataRecorder, error, func()) {
		switch testCase {
		case recorder_testing.RecorderErrorsOnUnavailableEndpointTestCase:
			err := recorder.ErrEndpointNotAvailable{Endpoint: "nowhere", Err: nil} // this is a special case because it's a mock
			return context.TODO(), (*recorder_testing.SimpleRecorder)(nil), err, func() {}

		case recorder_testing.RecorderBacksOffOnEndpointGoneTestCase:
			ctx, rec, err, teardown := setup(5)
			return ctx, rec, err, teardown

		default:
			return nil, nil, nil, nil
		}
	})
}
