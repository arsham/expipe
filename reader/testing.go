// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
)

const (
	// GenericReaderReceivesJobTestCase invokes TestGenericReaderReceivesJob test
	GenericReaderReceivesJobTestCase = iota
	// ReaderSendsResultTestCase invokes TestReaderSendsResult test
	ReaderSendsResultTestCase
	// ReaderReadsOnBufferedChanTestCase invokes TestReaderReadsOnBufferedChan test
	ReaderReadsOnBufferedChanTestCase
	// ReaderDrainsAfterClosingContextTestCase invokes TestReaderDrainsAfterClosingContext test
	ReaderDrainsAfterClosingContextTestCase
	// ReaderClosesTestCase invokes TestReaderCloses test
	ReaderClosesTestCase
	// ReaderClosesWithBufferedChansTestCase invokes TestReaderClosesWithBufferedChans test
	ReaderClosesWithBufferedChansTestCase
	// ReaderWithNoValidURLErrorsTestCase invokes TestReaderWithNoValidURLErrors test
	ReaderWithNoValidURLErrorsTestCase
	// ReaderErrorsOnEndpointDisapearsTestCase invokes TestReaderErrorsOnEndpointDisapears test
	ReaderErrorsOnEndpointDisapearsTestCase
)

// This file contains generic tests for various readers. You need to pass in your reader
// as a ready to use object, with all the necessary mocks, and these set of tests will do all
// the tests for you.
// IMPORTANT: you need to write the edge cases if they are not covered in this section.

// TestReaderEssentials runs all essential tests
func TestReaderEssentials(t *testing.T, setup func(testCase int) (red DataReader, errorChan chan communication.ErrorMessage, testMessage string, teardown func())) {
	t.Run("TestGenericReaderReceivesJob", func(t *testing.T) {
		red, _, _, _ := setup(GenericReaderReceivesJobTestCase)
		if red == nil {
			t.Fatal("You should implement GenericReaderReceivesJobTestCase")
		}
		testGenericReaderReceivesJob(t, red)
	})

	t.Run("TestReaderSendsResult", func(t *testing.T) {
		red, errorChan, testMessage, teardown := setup(ReaderSendsResultTestCase)
		if red == nil {
			t.Fatal("You should implement ReaderSendsResultTestCase")
		}
		defer teardown()
		testReaderSendsResult(t, red, errorChan, testMessage)
	})

	t.Run("TestReaderReadsOnBufferedChan", func(t *testing.T) {
		red, errorChan, testMessage, teardown := setup(ReaderReadsOnBufferedChanTestCase)
		if red == nil {
			t.Fatal("You should implement ReaderReadsOnBufferedChanTestCase")
		}
		defer teardown()
		testReaderReadsOnBufferedChan(t, red, errorChan, testMessage)
	})

	t.Run("TestReaderDrainsAfterClosingContext", func(t *testing.T) {
		red, errorChan, testMessage, teardown := setup(ReaderDrainsAfterClosingContextTestCase)
		if red == nil {
			t.Fatal("You should implement ReaderDrainsAfterClosingContextTestCase")
		}
		defer teardown()
		testReaderDrainsAfterClosingContext(t, red, errorChan, testMessage)
	})

	t.Run("TestReaderCloses", func(t *testing.T) {
		red, errorChan, testMessage, teardown := setup(ReaderClosesTestCase)
		if red == nil {
			t.Fatal("You should implement ReaderClosesTestCase")
		}
		defer teardown()
		testReaderCloses(t, red, errorChan, testMessage)
	})

	t.Run("TestReaderClosesWithBufferedChans", func(t *testing.T) {
		red, errorChan, testMessage, teardown := setup(ReaderClosesWithBufferedChansTestCase)
		if red == nil {
			t.Fatal("You should implement ReaderClosesWithBufferedChansTestCase")
		}
		defer teardown()
		testReaderClosesWithBufferedChans(t, red, errorChan, testMessage)
	})
}

// TestReaderEndpointManeuvers runs all tests regarding the endpoint changing state
func TestReaderEndpointManeuvers(t *testing.T, setup func(testCase int, endpoint string) (red DataReader, errorChan chan communication.ErrorMessage)) {
	t.Run("TestReaderErrorsOnEndpointDisapears", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		red, errorChan := setup(ReaderErrorsOnEndpointDisapearsTestCase, ts.URL)
		testReaderErrorsOnEndpointDisapears(t, ts, red, errorChan)
	})
}

type setupFunc func(
	name string,
	typeName string,
	endpoint string,
	jobChan chan context.Context,
	resultChan chan *ReadJobResult,
	interval time.Duration,
	timeout time.Duration,
) (DataReader, error)

// TestReaderConstruction runs all essential tests on object construction.
func TestReaderConstruction(t *testing.T, setup setupFunc) {
	name := "the name"
	typeName := "my type"
	endpoint := "http://127.0.0.1:9200"
	jobChan := make(chan context.Context)
	resultChan := make(chan *ReadJobResult)
	interval := time.Hour
	timeout := time.Hour

	testShowNotChangeTheInput(t, setup, name, typeName, endpoint, jobChan, resultChan, interval, timeout)
	testEndpointCheck(t, setup, name, typeName, endpoint, jobChan, resultChan, interval, timeout)

	// Name Check
	red, err := setup("", endpoint, typeName, jobChan, resultChan, interval, timeout)
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
	}
	if err != ErrEmptyName {
		t.Errorf("expected ErrEmptyName, got (%v)", err)
	}

	// TypeName Check
	red, err = setup(name, endpoint, "", jobChan, resultChan, interval, timeout)
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
	}
	if err != ErrEmptyTypeName {
		t.Errorf("expected ErrEmptyTypeName, got (%v)", err)
	}
}

func testShowNotChangeTheInput(t *testing.T, setup setupFunc, name string, typeName string, endpoint string, jobChan chan context.Context, resultChan chan *ReadJobResult, interval time.Duration, timeout time.Duration) {
	red, err := setup(name, endpoint, typeName, jobChan, resultChan, interval, timeout)
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}

	if red.Name() != name {
		t.Errorf("given name should not be changed: %v", red.Name())
	}
	if red.TypeName() != typeName {
		t.Errorf("given type name should not be changed: %v", red.TypeName())
	}
	if red.JobChan() != jobChan {
		t.Error("given job channel should not be changed")
	}
	if red.ResultChan() != resultChan {
		t.Error("given result channel should not be changed")
	}
	if red.Interval() != interval {
		t.Errorf("given interval should not be changed: %v", red.Timeout())
	}
	if red.Timeout() != timeout {
		t.Errorf("given timeout should not be changed: %v", red.Timeout())
	}
}

func testEndpointCheck(t *testing.T, setup setupFunc, name string, typeName string, endpoint string, jobChan chan context.Context, resultChan chan *ReadJobResult, interval time.Duration, timeout time.Duration) {

	red, err := setup(name, "", typeName, jobChan, resultChan, interval, timeout)
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
	}
	if err != ErrEmptyEndpoint {
		t.Errorf("expected ErrEmptyEndpoint, got (%v)", err)
	}

	invalidEndpoint := "this is invalid"
	red, err = setup(name, invalidEndpoint, typeName, jobChan, resultChan, interval, timeout)
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
	}
	if _, ok := err.(interface {
		InvalidEndpoint()
	}); !ok {
		t.Fatalf("expected ErrInvalidEndpoint, got (%v)", err)
	}
	if !strings.Contains(err.Error(), invalidEndpoint) {
		t.Errorf("expected (%s) be in the error message, got (%v)", invalidEndpoint, err)
	}

	unavailableEndpoint := "http://nowhere.localhost.localhost"
	red, err = setup(name, unavailableEndpoint, typeName, jobChan, resultChan, interval, timeout)
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
	}
	if err == nil {
		t.Fatal("expected ErrEndpointNotAvailable, got nil")
	}
	if _, ok := err.(interface {
		EndpointNotAvailable()
	}); !ok {
		t.Errorf("expected ErrEndpointNotAvailable, got (%v)", err)
	}
	if !strings.Contains(err.Error(), unavailableEndpoint) {
		t.Errorf("expected (%s) be in the error message, got (%v)", unavailableEndpoint, err)
	}

}

// testGenericReaderReceivesJob is a test helper to test the reader can receive jobs
func testGenericReaderReceivesJob(t *testing.T, red DataReader) {
	ctx := context.Background()
	stop := make(communication.StopChannel)
	red.Start(ctx, stop)

	select {
	case red.JobChan() <- communication.NewReadJob(ctx):
	case <-time.After(5 * time.Second):
		t.Error("expected the reader to receive the job, but it blocked")
	}
	done := make(chan struct{})
	stop <- done
	<-done
}

// testReaderSendsResult is a helper to test sending results
func testReaderSendsResult(t *testing.T, red DataReader, errorChan chan communication.ErrorMessage, testMessage string) {
	var res *ReadJobResult

	ctx := context.Background()
	stop := make(communication.StopChannel)
	red.Start(ctx, stop)

	red.JobChan() <- communication.NewReadJob(ctx)

	select {
	case err := <-errorChan:
		t.Errorf("didn't expect errors, got (%v)", err.Error())
	case <-time.After(20 * time.Millisecond):
	}

	select {
	case res = <-red.ResultChan():
	case <-time.After(5 * time.Second):
		t.Error("expected to receive a data back, nothing received")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Res)
	if buf.String() != testMessage {
		t.Errorf("want (%s), got (%s)", testMessage, buf.String())
	}
	done := make(chan struct{})
	stop <- done
	<-done
}

// testReaderReadsOnBufferedChan tests reading on buffered channels
func testReaderReadsOnBufferedChan(t *testing.T, red DataReader, errorChan chan communication.ErrorMessage, testMessage string) {
	var res *ReadJobResult
	ctx := context.Background()
	stop := make(communication.StopChannel)
	red.Start(ctx, stop)

	red.JobChan() <- communication.NewReadJob(ctx)

	select {
	case err := <-errorChan:
		t.Errorf("didn't expect errors, got (%v)", err.Error())
	case <-time.After(20 * time.Millisecond):
	}

	select {
	case res = <-red.ResultChan():
	case <-time.After(5 * time.Second):
		t.Error("expected to receive a data back, nothing received")
	}

	drained := false
	// Go is fast!
	for i := 0; i < 10; i++ {
		if len(red.JobChan()) == 0 {
			drained = true
			break
		}
		time.Sleep(10 * time.Millisecond)

	}
	if !drained {
		t.Errorf("expected to drain the jobChan, got (%d) left", len(red.JobChan()))
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Res)
	if buf.String() != testMessage {
		t.Errorf("want (%s), got (%s)", testMessage, buf.String())
	}
	done := make(chan struct{})
	stop <- done
	<-done
}

// testReaderDrainsAfterClosingContext tests the reader drains after closing the channel
func testReaderDrainsAfterClosingContext(t *testing.T, red DataReader, errorChan chan communication.ErrorMessage, testMessage string) {
	var res *ReadJobResult
	ctx := context.Background()
	stop := make(communication.StopChannel)
	red.Start(ctx, stop)

	red.JobChan() <- communication.NewReadJob(ctx)

	select {
	case err := <-errorChan:
		t.Errorf("didn't expect errors, got (%v)", err.Error())
	case <-time.After(20 * time.Millisecond):
	}

	select {
	case res = <-red.ResultChan():
	case <-time.After(5 * time.Second):
		t.Error("expected to receive a data back, nothing received")
	}

	drained := false
	// Go is fast!
	for i := 0; i < 10; i++ {
		if len(red.JobChan()) == 0 {
			drained = true
			break
		}
		time.Sleep(10 * time.Millisecond)

	}
	if !drained {
		t.Errorf("expected to drain the jobChan, got (%d) left", len(red.JobChan()))
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Res)
	if buf.String() != testMessage {
		t.Errorf("want (%s), got (%s)", testMessage, buf.String())
	}
	done := make(chan struct{})
	stop <- done
	<-done
}

// testReaderCloses tests the reader closes after finishing
func testReaderCloses(t *testing.T, red DataReader, errorChan chan communication.ErrorMessage, testMessage string) {
	var res *ReadJobResult
	ctx := context.Background()
	stop := make(communication.StopChannel)
	red.Start(ctx, stop)

	red.JobChan() <- communication.NewReadJob(ctx)
	res = <-red.ResultChan()
	defer res.Res.Close()
	done := make(chan struct{})
	stop <- done

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected to be done with the reader, but it blocked")
	}
}

// testReaderClosesWithBufferedChans tests the reader closes the with buffered channels
func testReaderClosesWithBufferedChans(t *testing.T, red DataReader, errorChan chan communication.ErrorMessage, testMessage string) {
	var res *ReadJobResult
	ctx := context.Background()
	stop := make(communication.StopChannel)
	red.Start(ctx, stop)

	red.JobChan() <- communication.NewReadJob(ctx)
	res = <-red.ResultChan()
	defer res.Res.Close()

	done := make(chan struct{})
	stop <- done
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected to be done with the reader, but it blocked")
	}
}

// testReaderErrorsOnEndpointDisapears is a helper to test the reader errors when the endpoint goes away.
func testReaderErrorsOnEndpointDisapears(t *testing.T, ts *httptest.Server, red DataReader, errorChan chan communication.ErrorMessage) {
	var res *ReadJobResult
	ctx := context.Background()
	stop := make(communication.StopChannel)
	red.Start(ctx, stop)
	ts.Close()
	job := communication.NewReadJob(ctx)
	red.JobChan() <- job

	select {
	case err := <-errorChan:
		switch interface{}(err).(type) {
		case communication.ErrorMessage:
		default:
			t.Errorf("want (communication.ErrorMessage), got (%v)", err.Error())
		}

		if err.ID != communication.JobValue(job) {
			t.Errorf("want (%v), got (%v)", communication.JobValue(job), err.ID)
		}
		if err.Name != red.Name() {
			t.Errorf("want (%s), got (%s)", red.Name(), err.Name)
		}
	case <-time.After(20 * time.Millisecond):
		t.Error("expected to receive an error, nothing received")
	}

	select {
	case res = <-red.ResultChan():
		t.Errorf("didn't expect to receive a data back, got (%v)", res)
	case <-time.After(10 * time.Millisecond):
	}
	done := make(chan struct{})
	stop <- done
	<-done
}
