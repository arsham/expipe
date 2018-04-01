// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expipe

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/reader"
	rdt "github.com/arsham/expipe/reader/testing"
	"github.com/arsham/expipe/recorder"
	rct "github.com/arsham/expipe/recorder/testing"
	"github.com/arsham/expipe/token"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus/hooks/test"
)

var (
	log        internal.FieldLogger
	testServer *httptest.Server
)

func init() {
	log = internal.DiscardLogger()
	testServer = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)
}

func withRecorder(ctx context.Context, log internal.FieldLogger) (*Engine, error) {
	rec, _ := rct.New(
		recorder.WithLogger(log),
		recorder.WithEndpoint(testServer.URL),
		recorder.WithName("recorder_test"),
		recorder.WithIndexName("indexName"),
		recorder.WithTimeout(time.Hour),
		recorder.WithBackoff(5),
	)

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

// EngineWithReadRecs creates an Engine instance with already set-up reader and
// recorders. The Engine's work starts from here by streaming all readers
// payloads to the recorder. Returns an error if there are recorders with
// the same name, or any of constructions results in errors.
//
// IMPORTANT: only use this for testing.
//
func EngineWithReadRecs(ctx context.Context, log internal.FieldLogger, rec recorder.DataRecorder, reds map[string]reader.DataReader) (*Engine, error) {
	var readerNames []string
	failedErrors := make(map[string]error)
	err := rec.Ping()
	if err != nil {
		return nil, PingError{rec.Name(): err}
	}

	readers := make(map[string]reader.DataReader)
	canDo := false
	i := 0
	for name, red := range reds {
		err := red.Ping()
		if err != nil {
			failedErrors[name] = err
			continue
		}
		readerNames = append(readerNames, name)
		readers[name] = red
		canDo = true
		i++
	}
	if !canDo {
		return nil, PingError(failedErrors)
	}
	// just to be cute
	engineName := fmt.Sprintf("( %s <-<< %s )", rec.Name(), strings.Join(readerNames, ","))
	log = log.WithField("engine", engineName)
	cl := &Engine{
		name:       engineName,
		ctx:        ctx,
		readerJobs: make(chan *reader.Result, len(reds)), // NOTE: increase this is as required
		recorder:   rec,
		readers:    readers,
		log:        log,
	}
	log.Debug("started the engine")
	return cl, nil
}

func TestEventLoopOneReaderSendsPayload(t *testing.T) {
	log := internal.DiscardLogger()
	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red, err := rdt.New(
		reader.WithLogger(internal.DiscardLogger()),
		reader.WithEndpoint(testServer.URL),
		reader.WithName("reader_name"),
		reader.WithTypeName("typeName"),
		reader.WithInterval(time.Millisecond),
		reader.WithTimeout(time.Second),
		reader.WithBackoff(5),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil): unexpected error occurred during reader creation", err)
	}
	if err := red.Ping(); err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}

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

	rec := e.recorder.(*rct.Recorder)
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if job.ID != jobID {
			t.Errorf("job.ID = (%s); want (%s)", job.ID, jobID)
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
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red1, err := rdt.New(
		reader.WithLogger(internal.DiscardLogger()),
		reader.WithEndpoint(testServer.URL),
		reader.WithName("reader_name"),
		reader.WithTypeName("typeName"),
		reader.WithInterval(time.Hour),
		reader.WithTimeout(time.Hour),
		reader.WithBackoff(5),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	if err = red1.Ping(); err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}

	red2, err := rdt.New(
		reader.WithLogger(internal.DiscardLogger()),
		reader.WithEndpoint(testServer.URL),
		reader.WithName("reader2_name"),
		reader.WithTypeName("typeName"),
		reader.WithInterval(time.Hour),
		reader.WithTimeout(time.Hour),
		reader.WithBackoff(5),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	if err = red2.Ping(); err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red1.ReadFunc = func(job *token.Context) (*reader.Result, error) { return nil, nil }
	red2.ReadFunc = func(job *token.Context) (*reader.Result, error) { return nil, nil }

	e.setReaders(map[string]reader.DataReader{red1.Name(): red1, red2.Name(): red2})

	rec := e.recorder.(*rct.Recorder)
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

func getReader(t *testing.T, name string, jobContent []byte) *rdt.Reader {
	red, err := rdt.New(
		reader.WithLogger(internal.DiscardLogger()),
		reader.WithEndpoint(testServer.URL),
		reader.WithName(name),
		reader.WithTypeName("typeName"),
		reader.WithInterval(time.Hour),
		reader.WithTimeout(time.Hour),
		reader.WithBackoff(5),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	if err = red.Ping(); err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}

	// testing engine send the payloads to the recorder
	red.ReadFunc = func(job *token.Context) (*reader.Result, error) {
		resp := &reader.Result{
			ID:       job.ID(),
			Content:  jobContent,
			TypeName: red.TypeName(),
			Mapper:   red.Mapper(),
		}
		return resp, nil
	}
	return red
}

// FIXME: break this test down
func TestEventLoopMultipleReadersSendPayload(t *testing.T) {
	log := internal.DiscardLogger()
	log.Level = internal.DebugLevel
	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}

	red1 := getReader(t, "reader1_name", []byte(`{"devil":666}`))
	red2 := getReader(t, "reader2_name", []byte(`{"beelzebub":666}`))
	e.setReaders(map[string]reader.DataReader{red1.Name(): red1, red2.Name(): red2})

	job1 := token.New(ctx)
	job2 := token.New(ctx)
	recorded := make(chan struct{})

	rec := e.recorder.(*rct.Recorder)
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if job.ID != job1.ID() && job.ID != job2.ID() {
			t.Errorf("job.ID = (%s); want one of (%s, %s)", job.ID, job1.ID(), job2.ID())
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
		if _, err := red1.Read(job1); err != nil {
			t.Error(err)
		}
		close(done1)
	}()
	go func() {
		if _, err := red1.Read(job2); err != nil {
			t.Error(err)
		}
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
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red, err := rdt.New(
		reader.WithLogger(internal.DiscardLogger()),
		reader.WithEndpoint(testServer.URL),
		reader.WithName("reader_name"),
		reader.WithTypeName("typeName"),
		reader.WithInterval(10*time.Millisecond),
		reader.WithTimeout(time.Second),
		reader.WithBackoff(5),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil): unexpected error occurred during reader creation", err)
	}
	if err = red.Ping(); err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}

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

func TestRemoveReader(t *testing.T) {
	endpoint := "http://localhost"
	r1, err := rdt.New(reader.WithName("r1"), reader.WithEndpoint(endpoint))
	if err != nil {
		t.Fatalf("rdt.New(name: r1): err = (%#v); want (nil)", err)
	}
	r2, err := rdt.New(reader.WithName("r2"), reader.WithEndpoint(endpoint))
	if err != nil {
		t.Fatalf("rdt.New(name: r2): err = (%#v); want (nil)", err)
	}
	r3, err := rdt.New(reader.WithName("r3"), reader.WithEndpoint(endpoint))
	if err != nil {
		t.Fatalf("rdt.New(name: r3): err = (%#v); want (nil)", err)
	}
	r4, err := rdt.New(reader.WithName("r4"), reader.WithEndpoint(endpoint))
	if err != nil {
		t.Fatalf("rdt.New(name: r4): err = (%#v); want (nil)", err)
	}
	e := &Engine{}
	e.setReaders(map[string]reader.DataReader{
		r1.Name(): r1,
		r2.Name(): r2,
		r3.Name(): r3,
		r4.Name(): r4,
	})
	e.removeReader(r1)
	if _, ok := e.readers[r1.Name()]; ok {
		t.Errorf("e.readers[r1.Name()]: didn't expect to have (%v) in the map", r1)
	}
	for _, r := range []reader.DataReader{r2, r3, r4} {
		if _, ok := e.readers[r.Name()]; !ok {
			t.Errorf("e.readers[r.Name()]: didn't expect it to remove (%v)", r)
		}
	}
}

func TestEventLoopCatchesReaderError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestEventLoopCatchesReaderError count in short mode")
	}
	t.Parallel()
	log, _ := test.NewNullLogger()

	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red, err := rdt.New(
		reader.WithLogger(internal.DiscardLogger()),
		reader.WithEndpoint(testServer.URL),
		reader.WithName("reader_name"),
		reader.WithTypeName("typeName"),
		reader.WithInterval(10*time.Millisecond),
		reader.WithTimeout(time.Second),
		reader.WithBackoff(5),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil): unexpected error occurred during reader creation", err)
	}
	if err = red.Ping(); err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}

	e.setReaders(map[string]reader.DataReader{red.Name(): red})

	errMsg := errMessage("an error happened")
	recorded := make(chan struct{})

	// Testing the engine catches errors
	red.ReadFunc = func(job *token.Context) (*reader.Result, error) {
		recorded <- struct{}{}
		return nil, errMsg
	}

	errChan := make(chan JobError)
	stop := make(chan struct{})
	e.issueReaderJob(red, errChan, stop)

	select {
	case <-recorded:
	case <-time.After(5 * time.Second):
		t.Error("expected to record, didn't happen")
	}
	cancel()
	select {
	case err := <-errChan:
		if errors.Cause(err.Err) != errMsg {
			t.Errorf("want (%v), got (%v)", errMsg, err.Err)
		}
	case <-time.After(5 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}
}
