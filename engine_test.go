// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/recorder"
)

// TODO: test engine closes readers when recorder goes out of scope

func TestNewWithReadRecorder(t *testing.T) {
	t.Parallel()
	log := lib.DiscardLogger()
	ctx := context.Background()

	jobChan := make(chan context.Context)
	errorChan := make(chan communication.ErrorMessage)
	resultChan := make(chan *reader.ReadJobResult)
	red, _ := reader.NewSimpleReader(log, reader.NewMockCtxReader("nowhere"), jobChan, resultChan, errorChan, "", "", time.Hour, time.Hour)

	payloadChan := make(chan *recorder.RecordJob)
	rec, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, errorChan, "a", "nowhere", "", time.Hour)

	e, err := expvastic.NewWithReadRecorder(ctx, log, errorChan, resultChan, rec, red)
	if err != expvastic.ErrEmptyRedName {
		t.Errorf("want ErrEmptyRedName, got (%v)", err)
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}

	red, _ = reader.NewSimpleReader(log, reader.NewMockCtxReader("nowhere"), jobChan, resultChan, errorChan, "a", "", time.Hour, time.Hour)
	red2, _ := reader.NewSimpleReader(log, reader.NewMockCtxReader("nowhere"), jobChan, resultChan, errorChan, "a", "", time.Hour, time.Hour)

	e, err = expvastic.NewWithReadRecorder(ctx, log, errorChan, resultChan, rec, red, red2)
	if err != expvastic.ErrDupRecName {
		t.Error("want error, got nil")
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}
}

func TestEngineSendJob(t *testing.T) {
	t.Parallel()
	var recorderID communication.JobID
	log := lib.DiscardLogger()
	ctx, cancel := context.WithCancel(context.Background())

	jobChan := make(chan context.Context)
	resultChan := make(chan *reader.ReadJobResult)
	errorChan := make(chan communication.ErrorMessage)

	ctxReader := reader.NewCtxReader("nowhere")
	red, _ := reader.NewSimpleReader(log, ctxReader, jobChan, resultChan, errorChan, "reader_example", "example_type", time.Hour, time.Hour)
	red.StartFunc = func(stop communication.StopChannel) {
		go func() {
			recorderID = communication.NewJobID()
			res := ioutil.NopCloser(bytes.NewBuffer([]byte(`{"devil":666}`)))
			resp := &reader.ReadJobResult{
				ID:       recorderID,
				Res:      res,
				TypeName: red.TypeName(),
				Mapper:   red.Mapper(),
			}
			resultChan <- resp
		}()
		go func() {
			s := <-stop
			s <- struct{}{}
		}()
	}

	payloadChan := make(chan *recorder.RecordJob)
	rec, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, errorChan, "recorder_example", "nowhere", "intexName", time.Hour)
	rec.StartFunc = func(stop communication.StopChannel) {
		go func() {
			recordedPayload := <-payloadChan

			if recordedPayload.ID != recorderID {
				t.Errorf("want (%d), got (%s)", recorderID, recordedPayload.ID)
			}
		}()
		go func() {
			s := <-stop
			s <- struct{}{}
		}()
	}

	e, err := expvastic.NewWithReadRecorder(ctx, log, errorChan, resultChan, rec, red)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()

	select {
	case err := <-errorChan:
		t.Fatalf("didn't expect errors, got (%v)", err)
	case <-time.After(5 * time.Millisecond): // Should be more than the interval, otherwise the response is not ready yet
	}
	// checking twice, non of reader and recorder should report any errors
	select {
	case err := <-errorChan:
		t.Fatalf("didn't expect errors, got (%v)", err)
	case <-time.After(5 * time.Millisecond): // Should be more than the interval, otherwise the response is not ready yet
	}

	cancel()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}
}

func TestEngineMultiReader(t *testing.T) {
	t.Parallel()
	count := 10
	log := lib.DiscardLogger()
	ctx, cancel := context.WithCancel(context.Background())
	IDs := make([]string, count)
	idChan := make(chan communication.JobID)
	for i := 0; i < count; i++ {
		id := communication.NewJobID()
		IDs[i] = id.String()
		go func(id communication.JobID) {
			idChan <- id
		}(id)
	}

	jobChan := make(chan context.Context)
	resultChan := make(chan *reader.ReadJobResult)
	errorChan := make(chan communication.ErrorMessage)
	payloadChan := make(chan *recorder.RecordJob)
	rec, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, errorChan, "recorder_example", "nowhere", "intexName", time.Hour)
	rec.StartFunc = func(stop communication.StopChannel) {
		go func() {
			recordedPayload := <-payloadChan

			if !lib.StringInSlice(recordedPayload.ID.String(), IDs) {
				t.Errorf("want once of (%s), got (%s)", strings.Join(IDs, ","), recordedPayload.ID)
			}

		}()
		go func() {
			s := <-stop
			s <- struct{}{}
		}()
	}

	reds := make([]reader.DataReader, count)
	for i := 0; i < count; i++ {

		ctxReader := reader.NewCtxReader("nowhere")
		name := fmt.Sprintf("reader_example_%d", i)
		red, _ := reader.NewSimpleReader(log, ctxReader, jobChan, resultChan, errorChan, name, "example_type", time.Hour, time.Hour)
		red.StartFunc = func(stop communication.StopChannel) {
			go func() {
				res := ioutil.NopCloser(bytes.NewBuffer([]byte(`{"devil":666}`)))
				resp := &reader.ReadJobResult{
					ID:       <-idChan,
					Res:      res,
					TypeName: red.TypeName(),
					Mapper:   red.Mapper(),
				}
				resultChan <- resp
			}()
			go func() {
				s := <-stop
				s <- struct{}{}
			}()

		}
		reds[i] = red
	}

	e, err := expvastic.NewWithReadRecorder(ctx, log, errorChan, resultChan, rec, reds...)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()

	for i := 0; i < count; i++ {
		select {
		case err := <-errorChan:
			t.Fatalf("didn't expect errors, got (%v)", err)
		case <-time.After(5 * time.Millisecond): // Should be more than the interval, otherwise the response is not ready yet
		}
	}

	cancel()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}
}

func TestEngineNewWithConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	log := lib.DiscardLogger()

	red, _ := reader.NewMockConfig("", "reader_example", log, "nowhere", "/still/nowhere", time.Hour, time.Hour, 1)
	rec, _ := recorder.NewMockConfig("reader_example", log, "nowhere", time.Hour, 1, "index")

	e, err := expvastic.NewWithConfig(ctx, log, 0, 0, 0, rec, red)
	if err != expvastic.ErrEmptyRedName {
		t.Error("want ErrEmptyRedName, got nil")
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}

	red, _ = reader.NewMockConfig("same_name_is_illegal", "reader_example", log, "nowhere", "/still/nowhere", time.Hour, time.Hour, 1)
	red2, _ := reader.NewMockConfig("same_name_is_illegal", "reader_example", log, "nowhere", "/still/nowhere", time.Hour, time.Hour, 1)

	e, err = expvastic.NewWithConfig(ctx, log, 0, 0, 0, rec, red, red2)
	if err != expvastic.ErrDupRecName {
		t.Error("want error, got nil")
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}
}
