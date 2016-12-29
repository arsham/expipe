// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
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
)

// This file contains generic tests for various recorders. You need to pass in your recorder
// as a ready to use object, with all the necessary mocks, and these set of tests will do all
// the tests for you.
// Note 1: you need to write the edge cases if they are not covered in this section.
// Note 2: the recorder should not be started.

// TestRecorderEssentials runs all essential tests.
// The only case the error is needed is to check the endpoint on start up.
func TestRecorderEssentials(t *testing.T, setup func(testCase int) (ctx context.Context, rec DataRecorder, err error, errorChan chan communication.ErrorMessage, teardown func())) {
	t.Run("TestRecorderReceivesPayload", func(t *testing.T) {
		ctx, rec, _, _, teardown := setup(RecorderReceivesPayloadTestCase)
		if rec == nil {
			t.Fatal("You should implement RecorderReceivesPayloadTestCase")
		}
		defer teardown()
		testRecorderReceivesPayload(ctx, t, rec)
	})

	t.Run("TestRecorderSendsResult", func(t *testing.T) {
		ctx, rec, _, errorChan, teardown := setup(RecorderSendsResultTestCase)
		if rec == nil {
			t.Fatal("You should implement RecorderSendsResultTestCase")
		}
		defer teardown()
		testRecorderSendsResult(ctx, t, rec, errorChan)
	})

	t.Run("TestRecorderCloses", func(t *testing.T) {
		ctx, rec, _, errorChan, teardown := setup(RecorderClosesTestCase)
		if rec == nil {
			t.Fatal("You should implement RecorderClosesTestCase")
		}
		defer teardown()
		testRecorderCloses(ctx, t, rec, errorChan)
	})

	t.Run("TestRecorderErrorsOnUnavailableESServer", func(t *testing.T) {
		_, rec, err, _, _ := setup(RecorderErrorsOnUnavailableEndpointTestCase)
		if err == nil {
			t.Fatal("You should implement RecorderErrorsOnUnavailableEndpointTestCase")
		}
		testRecorderErrorsOnUnavailableEndpoint(t, rec, err)
	})
}

func isTravis() bool {
	return os.Getenv("TRAVIS") != ""
}

type setupFunc func(
	payloadChan chan *RecordJob,
	name,
	indexName string,
	timeout time.Duration,
) DataRecorder

// TestRecorderConstruction runs all essential tests on object construction.
func TestRecorderConstruction(t *testing.T, setup setupFunc) {
	payloadChan := make(chan *RecordJob)
	name := "the name"
	indexName := "index_name"

	timeout := 10 * time.Millisecond
	if isTravis() {
		timeout = 10 * time.Second
	}

	rec := setup(payloadChan, name, indexName, timeout)

	if rec.PayloadChan() != payloadChan {
		t.Error("given payload channel should not be changed")
	}
	if rec.Name() != name {
		t.Errorf("given name should not be changed: %v", rec.Name())
	}
	if rec.IndexName() != indexName {
		t.Errorf("given index name should not be changed: %v", rec.IndexName())
	}
	if rec.Timeout() != timeout {
		t.Errorf("given timeout should not be changed: %v", rec.Timeout())
	}
}

// testRecorderReceivesPayload tests the recorder receives the payload correctly.
func testRecorderReceivesPayload(ctx context.Context, t *testing.T, rec DataRecorder) {
	stop := make(communication.StopChannel)
	rec.Start(ctx, stop)

	p := datatype.NewContainer([]datatype.DataType{})
	payload := &RecordJob{
		ID:        communication.NewJobID(),
		Ctx:       ctx,
		Payload:   p,
		IndexName: "my index",
		TypeName:  "my type",
		Time:      time.Now(),
	}
	select {
	case rec.PayloadChan() <- payload:
	case <-time.After(5 * time.Second):
		t.Error("expected the recorder to receive the payload, but it blocked")
	}
	done := make(chan struct{})
	stop <- done
	<-done

}

// testRecorderSendsResult tests the recorder send the results to the endpoint.
func testRecorderSendsResult(ctx context.Context, t *testing.T, rec DataRecorder, errorChan chan communication.ErrorMessage) {
	stop := make(communication.StopChannel)
	rec.Start(ctx, stop)

	p := datatype.NewContainer([]datatype.DataType{&datatype.StringType{Key: "test", Value: "test"}})
	payload := &RecordJob{
		ID:        communication.NewJobID(),
		Ctx:       ctx,
		Payload:   p,
		IndexName: "my index",
		TypeName:  "my type",
		Time:      time.Now(),
	}
	rec.PayloadChan() <- payload

	select {
	case err := <-errorChan:
		if err.Err != nil {
			t.Errorf("want (nil), got (%v)", err)
		}
	case <-time.After(20 * time.Millisecond):
	}
	done := make(chan struct{})
	stop <- done
	<-done
}

// testRecorderErrorsOnUnavailableEndpoint tests the recorder errors for bad URL.
func testRecorderErrorsOnUnavailableEndpoint(t *testing.T, rec DataRecorder, err error) {
	if err == nil {
		t.Error("want error, got nil")
	}
	if _, ok := err.(interface {
		EndpointNotAvailable()
	}); !ok {
		t.Errorf("want EndpointNotAvailable, got (%v)", err)
	}
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("want (nil), got (%v)", rec)
	}
}

// testRecorderCloses tests the recorder closes the done channel.
func testRecorderCloses(ctx context.Context, t *testing.T, rec DataRecorder, errorChan chan communication.ErrorMessage) {
	stop := make(communication.StopChannel)
	rec.Start(ctx, stop)

	done := make(chan struct{})
	stop <- done

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected the recorder to quit working")
	}
}
