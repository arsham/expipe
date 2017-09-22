// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/arsham/expvastic/internal"
	"github.com/arsham/expvastic/internal/token"
	"github.com/arsham/expvastic/reader"
	reader_test "github.com/arsham/expvastic/reader/testing"
	"github.com/arsham/expvastic/recorder"
	recorder_testing "github.com/arsham/expvastic/recorder/testing"
	"github.com/pkg/errors"
)

var (
	log        internal.FieldLogger
	testServer *httptest.Server
)

func init() {
	log = internal.DiscardLogger()
	testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
}

func withRecorder(ctx context.Context, log internal.FieldLogger) (*Engine, error) {
	rec, _ := recorder_testing.New(ctx, log, "recorder_test", testServer.URL, "indexName", time.Hour, 5)
	err := rec.Ping()
	if err != nil {
		return nil, err
	}
	return &Engine{
		name:       "test_engine",
		ctx:        ctx,
		log:        log,
		recorder:   rec,
		readerJobs: make(chan *reader.Result),
	}, nil
}

// setReaders is used only in tests.
func (e *Engine) setReaders(readers map[string]reader.DataReader) {
	e.redmu.Lock()
	defer e.redmu.Unlock()
	e.readers = readers
}

func TestEventLoopOneReaderSendsPayload(t *testing.T) {
	log := internal.DiscardLogger()

	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	red, err := reader_test.New(internal.DiscardLogger(), testServer.URL, "reader_name", "typeName", time.Millisecond, time.Millisecond, 5)
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}
	red.Ping()
	e.setReaders(map[string]reader.DataReader{red.Name(): red})
	job := token.New(ctx)
	jobID := job.ID()
	recorded := make(chan struct{})

	// testing engine send the payload to the recorder
	red.ReadFunc = func(job *token.Context) (*reader.Result, error) {
		resp := &reader.Result{
			ID:       jobID,
			Content:  []byte(`{"devil":666}`),
			TypeName: red.TypeName(),
			Mapper:   red.Mapper(),
		}
		return resp, nil
	}

	rec := e.recorder.(*recorder_testing.Recorder)
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if job.ID != jobID {
			t.Errorf("want (%s), got (%s)", jobID, job.ID)
		}
		recorded <- struct{}{}
		return nil
	}

	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()

	select {
	case <-recorded:
		cancel()
	case <-time.After(5 * time.Second):
		cancel()
		t.Error("expected to record, didn't happen")
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}
}

func TestEventLoopRecorderGoesOutOfScope(t *testing.T) {
	log := internal.DiscardLogger()
	log.Level = internal.DebugLevel

	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	red1, err := reader_test.New(internal.DiscardLogger(), testServer.URL, "reader_name", "typeName", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	red1.Ping()

	red2, _ := reader_test.New(internal.DiscardLogger(), testServer.URL, "reader2_name", "typeName", time.Hour, time.Hour, 5)
	red2.Ping()
	red1.ReadFunc = func(job *token.Context) (*reader.Result, error) { return nil, nil }
	red2.ReadFunc = func(job *token.Context) (*reader.Result, error) { return nil, nil }

	e.setReaders(map[string]reader.DataReader{red1.Name(): red1, red2.Name(): red2})

	rec := e.recorder.(*recorder_testing.Recorder)
	rec.RecordFunc = func(context.Context, *recorder.Job) error { return nil }

	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}
}

func TestEventLoopMultipleReadersSendPayload(t *testing.T) {
	log := internal.DiscardLogger()
	log.Level = internal.DebugLevel

	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	red1, err := reader_test.New(internal.DiscardLogger(), testServer.URL, "reader1_name", "typeName", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	red1.Ping()
	red2, err := reader_test.New(internal.DiscardLogger(), testServer.URL, "reader2_name", "typeName", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	red2.Ping()
	e.setReaders(map[string]reader.DataReader{red1.Name(): red1, red2.Name(): red2})

	job1 := token.New(ctx)
	job2 := token.New(ctx)
	recorded := make(chan struct{})

	// testing engine send the payloads to the recorder
	red1.ReadFunc = func(job *token.Context) (*reader.Result, error) {
		resp := &reader.Result{
			ID:       job.ID(),
			Content:  []byte(`{"devil":666}`),
			TypeName: red1.TypeName(),
			Mapper:   red1.Mapper(),
		}
		return resp, nil
	}

	red2.ReadFunc = func(job *token.Context) (*reader.Result, error) {
		resp := &reader.Result{
			ID:       job.ID(),
			Content:  []byte(`{"beelzebub":666}`),
			TypeName: red2.TypeName(),
			Mapper:   red2.Mapper(),
		}
		return resp, nil
	}

	rec := e.recorder.(*recorder_testing.Recorder)
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if job.ID != job1.ID() && job.ID != job2.ID() {
			t.Errorf("want one of (%s, %s), got (%s)", job1.ID(), job2.ID(), job.ID)
		}
		if job.ID != job1.ID() && job.ID != job2.ID() {
			t.Errorf("want one of (%s, %s), got (%s)", job1.ID(), job2.ID(), job.ID)
		}
		recorded <- struct{}{}
		recorded <- struct{}{}
		return nil
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		e.Start()
		wg.Done()
	}()
	done1 := make(chan struct{})
	done2 := make(chan struct{})
	go func() {
		red1.Read(job1)
		close(done1)
	}()
	go func() {
		red1.Read(job2)
		close(done2)
	}()

	select {
	case <-done1:
	case <-time.After(5 * time.Second):
		t.Error("expected red1 to record, didn't happen")
	}
	select {
	case <-done1:
	case <-time.After(5 * time.Second):
		t.Error("expected red1 to record, didn't happen")
	}

	cancel()
	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}

}

func TestStartReadersTicking(t *testing.T) {
	log := internal.DiscardLogger()

	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	red, err := reader_test.New(internal.DiscardLogger(), testServer.URL, "reader_name", "typeName", 10*time.Millisecond, 10*time.Millisecond, 5)
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}
	red.Ping()
	e.setReaders(map[string]reader.DataReader{red.Name(): red})

	recorded := make(chan struct{})

	// Testing the engine ticks and sends a job request to the reader
	// There is no need for the actual job
	red.ReadFunc = func(*token.Context) (*reader.Result, error) {
		recorded <- struct{}{} // important, otherwise the test might not be valid
		return nil, errors.New("blah blah")
	}

	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()

	select {
	case <-recorded:
	case <-time.After(2 * time.Second):
		t.Error("expected to record, didn't happen")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}
}
