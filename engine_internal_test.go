// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/test"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	reader_test "github.com/arsham/expvastic/reader/testing"
	"github.com/arsham/expvastic/recorder"
	recorder_testing "github.com/arsham/expvastic/recorder/testing"
	"github.com/arsham/expvastic/token"
)

var (
	log        logrus.FieldLogger
	testServer *httptest.Server
)

func init() {
	log = lib.DiscardLogger()
	testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	// testServer.Close()
}

type errMsg string

func (e errMsg) Error() string { return string(e) }

// inspectLogs checks if the niddle is found in the entries
// the entries might have been stacked, we need to iterate over.
func inspectLogs(entries []*logrus.Entry, niddle string) (all string, found bool) {
	var res []string

	for _, field := range entries {
		if strings.Contains(field.Message, niddle) {
			return "", true
		}
		res = append(res, field.Message)
	}
	return strings.Join(res, ", "), false
}

func withRecorder(ctx context.Context, log logrus.FieldLogger) (*Engine, error) {
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

func TestEventLoopCatchesReaderError(t *testing.T) {
	log, hook := test.NewNullLogger()
	log.Level = logrus.ErrorLevel

	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	red, err := reader_test.New(lib.DiscardLogger(), testServer.URL, "reader_name", "typeName", 10*time.Millisecond, 10*time.Millisecond, 5)
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}
	red.Ping()

	e.setReaders(map[string]reader.DataReader{red.Name(): red})

	errMsg := errMsg("an error happened")
	recorded := make(chan struct{})

	// Testing the engine catches errors
	red.ReadFunc = func(job *token.Context) (*reader.Result, error) {
		recorded <- struct{}{}
		return nil, errMsg
	}

	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()

	select {
	case <-recorded:
	case <-time.After(5 * time.Second):
		t.Error("expected to record, didn't happen")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}

	if _, found := inspectLogs(hook.Entries, errMsg.Error()); !found {
		// sometimes it takes time for logrus to register the error, trying again
		time.Sleep(500 * time.Millisecond)
		if all, found := inspectLogs(hook.Entries, errMsg.Error()); !found {
			t.Errorf("want (%s) in the error, got (%v)", errMsg.Error(), all)
		}
	}
}

func TestEventLoopOneReaderSendsPayload(t *testing.T) {
	log := lib.DiscardLogger()

	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	red, err := reader_test.New(lib.DiscardLogger(), testServer.URL, "reader_name", "typeName", time.Millisecond, time.Millisecond, 5)
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
	log, _ := test.NewNullLogger()
	log.Level = logrus.DebugLevel

	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	red1, err := reader_test.New(lib.DiscardLogger(), testServer.URL, "reader_name", "typeName", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	red1.Ping()

	red2, _ := reader_test.New(lib.DiscardLogger(), testServer.URL, "reader2_name", "typeName", time.Hour, time.Hour, 5)
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

func TestEventLoopClosingContext(t *testing.T) {
	log, hook := test.NewNullLogger()
	log.Level = logrus.DebugLevel

	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	red, err := reader_test.New(lib.DiscardLogger(), testServer.URL, "reader_name", "typeName", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}
	e.setReaders(map[string]reader.DataReader{red.Name(): red})

	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("expected the engine to quit gracefully")
	}

	if _, found := inspectLogs(hook.Entries, contextCanceled); !found {
		// sometimes it takes time for logrus to register the error, trying again
		time.Sleep(500 * time.Millisecond)
		if all, found := inspectLogs(hook.Entries, contextCanceled); !found {
			t.Errorf("want (%s) in the error, got (%v)", contextCanceled, all)
		}
	}
}

func TestEventLoopMultipleReadersSendPayload(t *testing.T) {
	log := lib.DiscardLogger()
	log.Level = logrus.DebugLevel

	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	red1, err := reader_test.New(lib.DiscardLogger(), testServer.URL, "reader1_name", "typeName", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	red1.Ping()
	red2, err := reader_test.New(lib.DiscardLogger(), testServer.URL, "reader2_name", "typeName", time.Hour, time.Hour, 5)
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
	log := lib.DiscardLogger()

	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	red, err := reader_test.New(lib.DiscardLogger(), testServer.URL, "reader_name", "typeName", 10*time.Millisecond, 10*time.Millisecond, 5)
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
		return nil, errMsg("blah blah")
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
