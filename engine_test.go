// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/recorder"
)

// TODO: test engine closes recorders when reader goes out of scope

func TestNewWithReadRecorder(t *testing.T) {
	log := lib.DiscardLogger()
	ctx := context.Background()

	jobChan := make(chan context.Context)
	errorChan := make(chan communication.ErrorMessage)
	resultChan := make(chan *reader.ReadJobResult)
	red, _ := reader.NewSimpleReader(log, reader.NewMockCtxReader("nowhere"), jobChan, resultChan, errorChan, "a", "", time.Hour, time.Hour)

	payloadChan := make(chan *recorder.RecordJob)
	rec, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, errorChan, "", "nowhere", "", time.Hour)

	e, err := expvastic.NewWithReadRecorder(ctx, log, 0, errorChan, red, rec)
	if err != expvastic.ErrEmptyRecName {
		t.Error("want ErrEmptyRecName, got nil")
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}

	rec, _ = recorder.NewSimpleRecorder(ctx, log, payloadChan, errorChan, "same_name_is_illegal", "nowhere", "", time.Hour)
	rec2, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, errorChan, "same_name_is_illegal", "nowhere", "", time.Hour)

	e, err = expvastic.NewWithReadRecorder(ctx, log, 0, errorChan, red, rec, rec2)
	if err != expvastic.ErrDupRecName {
		t.Error("want error, got nil")
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}
}

func TestEngineSendJob(t *testing.T) {
	var res *reader.ReadJobResult
	log := lib.DiscardLogger()
	ctx, cancel := context.WithCancel(context.Background())
	desire := `{"the key": "is the value!"}`
	recorded := make(chan struct{})

	redTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, desire)
	}))
	defer redTs.Close()

	recTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorded <- struct{}{}
	}))
	defer recTs.Close()

	jobChan := make(chan context.Context)
	resultChan := make(chan *reader.ReadJobResult)
	errorChan := make(chan communication.ErrorMessage)
	ctxReader := reader.NewCtxReader(redTs.URL)
	red, _ := reader.NewSimpleReader(log, ctxReader, jobChan, resultChan, errorChan, "reader_example", "example_type", time.Hour, time.Hour)
	redDone := red.Start(ctx)

	payloadChan := make(chan *recorder.RecordJob)
	rec, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, errorChan, "recorder_example", recTs.URL, "intexName", time.Hour)
	recDone := rec.Start(ctx)

	cl, err := expvastic.NewWithReadRecorder(ctx, log, 0, errorChan, red, rec)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	clDone := cl.Start()
	select {
	case red.JobChan() <- communication.NewReadJob(ctx):
	case <-time.After(time.Second):
		t.Error("expected the reader to recive the job, but it blocked")
	}

	select {
	case err := <-errorChan:
		t.Fatalf("didn't expect errors, got (%v)", err)
	case <-time.After(5 * time.Millisecond): // Should be more than the interval, otherwise the response is not ready yet
	}

	select {
	case res = <-red.ResultChan():
	case <-time.After(5 * time.Second): // Should be more than the interval, otherwise the response is not ready yet
		t.Error("expected to recive a data back, nothing recieved")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Res)
	if buf.String() != desire {
		t.Errorf("want (%s), got (%s)", desire, buf.String())
	}

	select {
	case <-recorded:
	case <-time.After(5 * time.Second):
		t.Errorf("recorder didn't record the request")
	}

	cancel()
	if _, ok := <-redDone; ok {
		t.Error("expected the channel to be closed")
	}
	if _, ok := <-recDone; ok {
		t.Error("expected the channel to be closed")
	}
	if v, ok := <-clDone; ok {
		t.Error("expected the channel to be closed", v)
	}
}

func testEngineMultiRecorder(t *testing.T) {
	var res *reader.ReadJobResult
	count := 10
	log := lib.DiscardLogger()
	ctx, cancel := context.WithCancel(context.Background())
	desire := `{"the key": "is the value!"}`
	recorded := make(chan struct{})

	redTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, desire)
	}))
	defer redTs.Close()

	jobChan := make(chan context.Context)
	resultChan := make(chan *reader.ReadJobResult)
	errorChan := make(chan communication.ErrorMessage)
	ctxReader := reader.NewCtxReader(redTs.URL)
	red, _ := reader.NewSimpleReader(log, ctxReader, jobChan, resultChan, errorChan, "reader_example", "example_type", time.Hour, time.Hour)
	redDone := red.Start(ctx)

	recs := make([]recorder.DataRecorder, count)
	recsDone := make([]<-chan struct{}, count)
	for i := 0; i < count; i++ {
		payloadChan := make(chan *recorder.RecordJob)
		name := fmt.Sprintf("recorder_example_%d", i)
		rec, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, errorChan, name, "does not matter", "intexName", time.Hour)

		rec.StartFunc = func() chan struct{} {
			done := make(chan struct{})
			go func(payloadChan chan *recorder.RecordJob) {
			LOOP:
				for {
					select {
					case <-payloadChan:
						recorded <- struct{}{}
					case <-ctx.Done():
						break LOOP
					}
				}
				close(done)
			}(payloadChan)
			return done
		}

		recs[i] = rec
		recsDone[i] = rec.Start(ctx)
	}

	cl, err := expvastic.NewWithReadRecorder(ctx, log, 0, errorChan, red, recs...)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	clDone := cl.Start()

	select {
	case red.JobChan() <- communication.NewReadJob(ctx):
	case <-time.After(time.Second):
		t.Error("expected the reader to recive the job, but it blocked")
	}

	select {
	case err := <-errorChan:
		t.Fatalf("didn't expect errors, got (%v)", err)
	case <-time.After(5 * time.Millisecond): // Should be more than the interval, otherwise the response is not ready yet
	}

	select {
	case res = <-red.ResultChan():
	case <-time.After(5 * time.Second): // Should be more than the interval, otherwise the response is not ready yet
		t.Error("expected to recive a data back, nothing recieved")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Res)
	if buf.String() != desire {
		t.Errorf("want (%s), got (%s)", desire, buf.String())
	}

	for i := 0; i < count; i++ {
		select {
		case <-recorded:
		case <-time.After(20 * time.Second):
			t.Errorf("recorder didn't record the request")
		}
	}

	cancel()
	if _, ok := <-redDone; ok {
		t.Error("expected the channel to be closed")
	}
	for _, done := range recsDone {
		select {
		case <-done:
		case <-time.After(20 * time.Second):
			t.Error("expected the recorder to finish")
		}
	}

	if v, ok := <-clDone; ok {
		t.Error("expected the channel to be closed", v)
	}
}

func TestEngineNewWithConfig(t *testing.T) {
	ctx := context.Background()
	log := lib.DiscardLogger()

	red, _ := reader.NewMockConfig("reader_example", "reader_example", log, "nowhere", "/still/nowhere", time.Hour, time.Hour, 1)
	rec, _ := recorder.NewMockConfig("", log, "nowhere", time.Hour, 1, "index")

	e, err := expvastic.NewWithConfig(ctx, log, 0, 0, 0, 0, red, rec)
	if err != expvastic.ErrEmptyRecName {
		t.Error("want ErrEmptyRecName, got nil")
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}

	rec, _ = recorder.NewMockConfig("same_name_is_illegal", log, "nowhere", time.Hour, 1, "index")
	rec2, _ := recorder.NewMockConfig("same_name_is_illegal", log, "nowhere", time.Hour, 1, "index")

	e, err = expvastic.NewWithConfig(ctx, log, 0, 0, 0, 0, red, rec, rec2)
	if err != expvastic.ErrDupRecName {
		t.Error("want error, got nil")
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}
}
