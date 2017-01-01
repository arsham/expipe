// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/arsham/expvastic/recorder"
)

const (
	// RecorderReceivesPayloadTestCase is for invoking TestRecorderReceivesPayload test
	RecorderReceivesPayloadTestCase = iota

	// RecorderSendsResultTestCase is for invoking TestRecorderSendsResult test
	RecorderSendsResultTestCase

	// RecorderClosesTestCase is for invoking TestRecorderCloses test
	RecorderClosesTestCase

	// RecorderErrorsOnUnavailableEndpointTestCase is for invoking TestRecorderErrorsOnUnavailableEndpoint test
	RecorderErrorsOnUnavailableEndpointTestCase

	// RecorderBacksOffOnEndpointGoneTestCase invokes TestRecorderBacksOffOnEndpointGone test
	RecorderBacksOffOnEndpointGoneTestCase

	// RecorderConstructionCasesTestCase is for invoking TestRecorderConstructionCases test
	RecorderConstructionCasesTestCase

	// RecorderErrorsOnInvalidEndpointTestCase is for invoking TestRecorderErrorsOnInvalidEndpoint test
	RecorderErrorsOnInvalidEndpointTestCase
)

// This file contains generic tests for various recorders. You need to pass in your recorder
// as a ready to use object, with all the necessary mocks, and these set of tests will do all
// the tests for you.
// Note 1: you need to write the edge cases if they are not covered in this section.
// Note 2: the recorder should not be started.

func isTravis() bool {
	return os.Getenv("TRAVIS") != ""
}

type setupFunc func(
	testCase int,
	name,
	endpoint,
	indexName string,
	timeout time.Duration,
	backoff int,
) (recorder.DataRecorder, error)

// TestRecorderConstruction runs all essential tests on object construction.
func TestRecorderConstruction(t *testing.T, setup setupFunc) {
	t.Run("TestRecorderConstructionCases", func(t *testing.T) {
		testRecorderConstructionCases(t, setup)
	})

	t.Run("TestRecorderErrorsOnInvalidEndpoint", func(t *testing.T) {
		testRecorderErrorsOnInvalidEndpoint(t, setup)
	})
}

// TestRecorderCommunication runs all essential tests.
func TestRecorderCommunication(t *testing.T, setup func(testCase int) (ctx context.Context, rec recorder.DataRecorder, err error, teardown func())) {
	t.Run("TestRecorderReceivesPayload", func(t *testing.T) {
		ctx, rec, _, teardown := setup(RecorderReceivesPayloadTestCase)
		if rec == nil {
			t.Fatal("You should implement RecorderReceivesPayloadTestCase")
		}
		defer teardown()
		testRecorderReceivesPayload(ctx, t, rec)
	})

	t.Run("TestRecorderSendsResult", func(t *testing.T) {
		ctx, rec, _, teardown := setup(RecorderSendsResultTestCase)
		if rec == nil {
			t.Fatal("You should implement RecorderSendsResultTestCase")
		}
		defer teardown()
		testRecorderSendsResult(ctx, t, rec)
	})
}

// TestRecorderEndpointManeuvers runs all tests regarding the endpoint changing state.
func TestRecorderEndpointManeuvers(t *testing.T, setup func(testCase int) (ctx context.Context, rec recorder.DataRecorder, err error, teardown func())) {
	t.Run("TestRecorderErrorsOnUnavailableESServer", func(t *testing.T) {
		_, rec, err, _ := setup(RecorderErrorsOnUnavailableEndpointTestCase)
		if rec == nil {
			t.Fatal("You should implement RecorderErrorsOnUnavailableEndpointTestCase")
		}
		testRecorderErrorsOnUnavailableEndpoint(t, rec, err)
	})

	t.Run("TestRecorderBacksOffOnEndpointGone", func(t *testing.T) {
		_, rec, _, teardown := setup(RecorderBacksOffOnEndpointGoneTestCase)
		if rec == nil {
			t.Fatal("You should implement RecorderBacksOffOnEndpointGoneTestCase")
		}
		testRecorderBacksOffOnEndpointGone(t, rec, teardown)
	})
}
