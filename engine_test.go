// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/recorder"
)

// TODO: test engine closes all recorders
// TODO: test engine closes recorders when reader goes out of scope

func TestNewWithReadRecorder(t *testing.T) {
	log := lib.DiscardLogger()
	ctx := context.Background()

	jobChan := make(chan context.Context)
	resultChan := make(chan *reader.ReadJobResult)
	red, _ := reader.NewSimpleReader(log, reader.NewMockCtxReader("nowhere"), jobChan, resultChan, "a", "", time.Millisecond, time.Millisecond)

	payloadChan := make(chan *recorder.RecordJob)
	rec, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, "", "nowhere", "", time.Millisecond, time.Millisecond)

	e, err := expvastic.NewWithReadRecorder(ctx, log, 0, red, rec)
	if err != expvastic.ErrEmptyRecName {
		t.Error("want ErrEmptyRecName, got nil")
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}

	rec, _ = recorder.NewSimpleRecorder(ctx, log, payloadChan, "same_name_is_illegal", "nowhere", "", time.Millisecond, time.Millisecond)
	rec2, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, "same_name_is_illegal", "nowhere", "", time.Millisecond, time.Millisecond)

	e, err = expvastic.NewWithReadRecorder(ctx, log, 0, red, rec, rec2)
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
	ctxReader := reader.NewCtxReader(redTs.URL)
	red, _ := reader.NewSimpleReader(log, ctxReader, jobChan, resultChan, "reader_example", "example_type", time.Millisecond, time.Millisecond)
	redDone := red.Start(ctx)

	payloadChan := make(chan *recorder.RecordJob)
	rec, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, "recorder_example", recTs.URL, "intexName", time.Millisecond, time.Millisecond)
	recDone := rec.Start(ctx)

	cl, err := expvastic.NewWithReadRecorder(ctx, log, 0, red, rec)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	clDone := cl.Start()

	select {
	case red.JobChan() <- ctx:
	case <-time.After(time.Second):
		t.Error("expected the reader to recive the job, but it blocked")
	}

	select {
	case res = <-red.ResultChan():
		if res.Err != nil {
			t.Fatalf("want (nil), got (%v)", res.Err)
		}
	case <-time.After(5 * time.Second): // Should be more than the interval, otherwise the response is not ready yet
		//TODO: investigate
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

func TestEngineMultiRecorder(t *testing.T) {
	log := lib.DiscardLogger()
	ctx, cancel := context.WithCancel(context.Background())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()

	jobChan := make(chan context.Context)
	resultChan := make(chan *reader.ReadJobResult)
	red, _ := reader.NewSimpleReader(log, &reader.MockCtxReader{}, jobChan, resultChan, "reader_example", "example_type", time.Second, time.Second)
	red.StartFunc = func() chan struct{} {
		done := make(chan struct{})
		go func() {
			<-red.JobChan()
			res := &reader.ReadJobResult{
				Time:     time.Now(),
				Res:      ioutil.NopCloser(bytes.NewBufferString("")),
				Err:      nil,
				TypeName: "example_type",
			}
			red.ResultChan() <- res
			<-ctx.Done()
			close(done)
			return
		}()
		return done
	}

	length := 10
	recorders := make([]recorder.DataRecorder, length)
	results := make(chan struct{}, 20)
	for i := 0; i < length; i++ {
		name := fmt.Sprintf("rec_%d", i)
		payloadChan := make(chan *recorder.RecordJob)
		rec, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, name, ts.URL, "intexName", time.Second, time.Second)

		go func(rec *recorder.SimpleRecorder) {
			job := make(chan *recorder.RecordJob)
			done := make(chan struct{})

			rec.Smu.Lock()
			rec.StartFunc = func() chan struct{} {
				return done
			}
			rec.Smu.Unlock()

			rec.Pmu.Lock()
			rec.PayloadChanFunc = func() chan *recorder.RecordJob {
				return job
			}
			rec.Pmu.Unlock()

			j := <-job
			j.Err <- nil
			results <- struct{}{}
			close(done)
		}(rec)

		recorders[i] = rec
	}

	cl, err := expvastic.NewWithReadRecorder(ctx, log, 0, red, recorders...)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	clDone := cl.Start()

	red.JobChan() <- ctx
	for i := 0; i < length; i++ {
		<-results
	}

	if len(results) != 0 {
		t.Errorf("want (%d) results, got (%d)", length, len(results)+length)
	}
	cancel()
	<-clDone
}

func TestEngineNewWithConfig(t *testing.T) {
	ctx := context.Background()
	log := lib.DiscardLogger()

	red, _ := reader.NewMockConfig("reader_example", "reader_example", log, "nowhere", "/still/nowhere", time.Millisecond, time.Millisecond, 1)
	rec, _ := recorder.NewMockConfig("", log, "nowhere", time.Millisecond, time.Millisecond, 1, "index")

	e, err := expvastic.NewWithConfig(ctx, log, 0, 0, 0, 0, red, rec)
	if err != expvastic.ErrEmptyRecName {
		t.Error("want ErrEmptyRecName, got nil")
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}

	rec, _ = recorder.NewMockConfig("same_name_is_illegal", log, "nowhere", time.Millisecond, time.Millisecond, 1, "index")
	rec2, _ := recorder.NewMockConfig("same_name_is_illegal", log, "nowhere", time.Millisecond, time.Millisecond, 1, "index")

	e, err = expvastic.NewWithConfig(ctx, log, 0, 0, 0, 0, red, rec, rec2)
	if err != expvastic.ErrDupRecName {
		t.Error("want error, got nil")
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}
}
