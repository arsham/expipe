// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/arsham/expipe/engine"
	"github.com/arsham/expipe/reader"
	rdt "github.com/arsham/expipe/reader/testing"
	"github.com/arsham/expipe/recorder"
	rct "github.com/arsham/expipe/recorder/testing"
	"github.com/arsham/expipe/tools"
	"github.com/arsham/expipe/tools/token"

	"github.com/pkg/errors"
)

var (
	log        tools.FieldLogger
	testServer *httptest.Server
)

func TestMain(m *testing.M) {
	log = tools.DiscardLogger()
	testServer = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)
	exitCode := m.Run()
	testServer.Close()
	os.Exit(exitCode)
}

func sampleReader(t *testing.T, ping bool) *rdt.Reader {
	red, err := rdt.New(
		reader.WithLogger(log),
		reader.WithEndpoint(testServer.URL),
		reader.WithName("red_name"),
		reader.WithTypeName("type_name"),
		reader.WithInterval(time.Second),
		reader.WithTimeout(time.Second),
		reader.WithBackoff(5),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	if !ping {
		return red
	}
	if err := red.Ping(); err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	return red
}

func sampleRecorder(t *testing.T, ping bool) *rct.Recorder {
	rec, err := rct.New(
		recorder.WithLogger(log),
		recorder.WithEndpoint(testServer.URL),
		recorder.WithName("rec_name"),
		recorder.WithIndexName("index_name"),
		recorder.WithTimeout(time.Second),
		recorder.WithBackoff(5),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	if !ping {
		return rec
	}
	if err := rec.Ping(); err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	return rec
}

func TestNewWithReadRecorder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	rec := sampleRecorder(t, true)
	red := sampleReader(t, true)
	red2 := sampleReader(t, true)
	red2.SetName("d")

	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithReaders(red),
		engine.WithLogger(log),
		engine.WithRecorder(rec),
	)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if e == nil {
		t.Error("e = (nil); want (Engine)")
	}
}

func TestEngineSendJob(t *testing.T) {
	t.Parallel()
	var recorderID token.ID
	ctx, cancel := context.WithCancel(context.Background())
	red := sampleReader(t, true)
	red.ReadFunc = func(job *token.Context) (*reader.Result, error) {
		recorderID = token.NewUID()
		resp := &reader.Result{
			ID:       recorderID,
			Content:  []byte(`{"devil":666}`),
			TypeName: red.TypeName(),
			Mapper:   red.Mapper(),
		}
		return resp, nil
	}

	rec := sampleRecorder(t, true)
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if job.ID != recorderID {
			t.Errorf("want (%d), got (%s)", recorderID, job.ID)
		}
		return nil
	}

	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithReaders(red),
		engine.WithLogger(log),
		engine.WithRecorder(rec),
	)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	done := make(chan struct{})
	go func() {
		engine.Start(e)
		done <- struct{}{}
	}()

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
	ctx, cancel := context.WithCancel(context.Background())
	IDs := make([]string, count)
	idChan := make(chan token.ID)
	for i := 0; i < count; i++ {
		id := token.NewUID()
		IDs[i] = id.String()
		go func(id token.ID) {
			idChan <- id
		}(id)
	}
	rec := sampleRecorder(t, false)
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if !tools.StringInSlice(job.ID.String(), IDs) {
			t.Errorf("job.ID = (%s); want once of (%s)", job.ID, strings.Join(IDs, ","))
		}
		return nil
	}

	reds := make([]reader.DataReader, count)
	for i := 0; i < count; i++ {
		var red *rdt.Reader
		name := fmt.Sprintf("reader_example_%d", i)
		red = sampleReader(t, false)
		red.SetName(name)
		red.ReadFunc = func(job *token.Context) (*reader.Result, error) {
			resp := &reader.Result{
				ID:       <-idChan,
				Content:  []byte(`{"devil":666}`),
				TypeName: red.TypeName(),
				Mapper:   red.Mapper(),
			}
			return resp, nil
		}
		reds[i] = red
	}

	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithReaders(reds...),
		engine.WithLogger(log),
		engine.WithRecorder(rec),
	)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	done := make(chan struct{})
	go func() {
		engine.Start(e)
		done <- struct{}{}
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}
}

func TestEngineErrorsIfReaderNotPinged(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	redServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	recServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer recServer.Close()
	redServer.Close() // making sure no one else is got this random port at this time

	rec := sampleRecorder(t, false)
	rec.SetEndpoint(recServer.URL)

	red := sampleReader(t, false)
	red.SetEndpoint(redServer.URL)

	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithReaders(red),
		engine.WithLogger(log),
		engine.WithRecorder(rec),
	)
	if err == nil {
		t.Error("err = (nil); want (PingError)")
	}

	if _, ok := errors.Cause(err).(engine.PingError); !ok {
		t.Errorf("err = (%#v); want (PingError)", err)
	}
	if e != nil {
		t.Errorf("e = (%v); want (nil)", e)
	}
}

func TestEngineErrorsIfRecorderNotPinged(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	redServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	recServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	recServer.Close() // making sure no one else is got this random port at this time
	defer redServer.Close()

	rec := sampleRecorder(t, false)
	rec.SetEndpoint(recServer.URL)

	red := sampleReader(t, false)

	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithReaders(red),
		engine.WithLogger(log),
		engine.WithRecorder(rec),
	)
	if err == nil {
		t.Error("err = (nil); want (PingError)")
	}

	if _, ok := errors.Cause(err).(engine.PingError); !ok {
		t.Errorf("err = (%v); want (PingError)", err)
	}
	if e != nil {
		t.Errorf("e = (%v); want (nil)", e)
	}
}

func TestEngineOnlyErrorsIfNoneOfReadersPinged(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	deadServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	liveServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer liveServer.Close()
	deadServer.Close() // making sure no one else is got this random port at this time

	rec := sampleRecorder(t, false)
	red1 := sampleReader(t, false)
	red1.SetEndpoint(liveServer.URL)

	red2 := sampleReader(t, false)
	red2.SetEndpoint(deadServer.URL)
	red2.SetName("b")
	red2.SetTypeName("ddb")

	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithReaders(red1, red2),
		engine.WithLogger(log),
		engine.WithRecorder(rec),
	)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if e == nil {
		t.Error("e = (nil); want (Engine)")
	}

	// now the engine should error
	red1 = sampleReader(t, false)
	red1.SetEndpoint(deadServer.URL)
	red1.SetName("a")
	red1.SetTypeName("ddc")

	err = engine.WithReaders(red1, red2)(e)
	if err == nil {
		t.Error("err = (nil); want (PingError)")
	}
	if _, ok := errors.Cause(err).(engine.PingError); !ok {
		t.Errorf("err = (%v); want (PingError)", err)
	}
}

// FIXME: break this test down
func TestEngineShutsDownOnAllReadersGoOutOfScope(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestEngineShutsDownOnAllReadersGoOutOfScope count in short mode")
	}
	t.Parallel()
	stopReader1 := uint32(0)
	stopReader2 := uint32(0)
	readerInterval := time.Millisecond * 10

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	red1 := sampleReader(t, false)
	red1.SetName("reader1_example")

	red2 := sampleReader(t, false)
	red2.SetName("reader2_example")

	red1.ReadFunc = func(job *token.Context) (*reader.Result, error) {
		if atomic.LoadUint32(&stopReader1) > 0 {
			return nil, reader.ErrBackoffExceeded
		}
		resp := &reader.Result{
			ID:       token.NewUID(),
			Content:  []byte(`{"devil":666}`),
			TypeName: red1.TypeName(),
			Mapper:   red1.Mapper(),
		}
		return resp, nil
	}

	red2.ReadFunc = func(job *token.Context) (*reader.Result, error) {
		if atomic.LoadUint32(&stopReader2) > 0 {
			return nil, reader.ErrBackoffExceeded
		}
		resp := &reader.Result{
			ID:       token.NewUID(),
			Content:  []byte(`{"devil":666}`),
			TypeName: red2.TypeName(),
			Mapper:   red2.Mapper(),
		}
		return resp, nil
	}

	rec := sampleRecorder(t, false)
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error { return nil }

	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithReaders(red1, red2),
		engine.WithLogger(log),
		engine.WithRecorder(rec),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}

	cleanExit := make(chan struct{})
	go func() {
		engine.Start(e)
		cleanExit <- struct{}{}
	}()

	// check the engine is working correctly with one reader
	time.Sleep(readerInterval * 3) // making sure it reads at least once
	atomic.StoreUint32(&stopReader1, uint32(1))
	time.Sleep(readerInterval * 2) // making sure the engine is not falling over

	select {
	case <-cleanExit:
		t.Fatal("expected the engine continue")
	case <-time.After(readerInterval * 2):
	}

	time.Sleep(readerInterval * 2)
	atomic.StoreUint32(&stopReader2, uint32(1))

	select {
	case <-cleanExit:
	case <-time.After(5 * time.Second):
		t.Error("expected the engine to quit")
	}
}

func TestEngineShutsDownOnRecorderGoOutOfScope(t *testing.T) {
	t.Parallel()
	stopRecorder := uint32(0)
	readerInterval := time.Millisecond * 10
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	red := sampleReader(t, false)
	red.SetInterval(time.Millisecond * 50)

	red.ReadFunc = func(job *token.Context) (*reader.Result, error) {
		resp := &reader.Result{
			ID:       token.NewUID(),
			Content:  []byte(`{"devil":666}`),
			TypeName: red.TypeName(),
			Mapper:   red.Mapper(),
		}
		return resp, nil
	}

	rec := sampleRecorder(t, false)
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if atomic.LoadUint32(&stopRecorder) > 0 {
			return recorder.ErrBackoffExceeded
		}
		return nil
	}

	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithReaders(red),
		engine.WithLogger(log),
		engine.WithRecorder(rec),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}

	cleanExit := make(chan struct{})
	go func() {
		engine.Start(e)
		cleanExit <- struct{}{}
	}()

	// check the engine is working correctly with one reader
	time.Sleep(readerInterval * 3) // making sure it reads at least once
	atomic.StoreUint32(&stopRecorder, 1)
	time.Sleep(readerInterval * 2) // making sure the engine is not falling over

	select {
	case <-cleanExit:
	case <-time.After(5 * time.Second):
		t.Error("expected the engine to quit")
	}
}

func TestFailsOnNilReader(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	rec := sampleRecorder(t, false)

	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithLogger(log),
		engine.WithRecorder(rec),
	)
	if errors.Cause(err) != engine.ErrNoReader {
		t.Errorf("err = (%v); want (NoReaderError)", err)
	}
	if e != nil {
		t.Errorf("e = (%v); want (Engine)", e)
	}

	e, err = engine.New(
		engine.WithCtx(ctx),
		engine.WithLogger(log),
		engine.WithRecorder(rec),
	)
	if errors.Cause(err) != engine.ErrNoReader {
		t.Errorf("err = (%v); want (NoReaderError)", err)
	}
	if e != nil {
		t.Errorf("e = (%v); want (Engine)", e)
	}
}

func TestFailsOnNilRecorder(t *testing.T) {
	t.Parallel()
	e, err := engine.New(
		engine.WithRecorder(nil),
	)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if e != nil {
		t.Errorf("e = (%v); want (Engine)", e)
	}
}

func TestEngineFailsNoLog(t *testing.T) {
	t.Parallel()
	red := sampleReader(t, false)
	rec := sampleRecorder(t, false)

	e, err := engine.New(
		engine.WithRecorder(rec),
		engine.WithReaders(red),
		engine.WithCtx(context.Background()),
	)
	if errors.Cause(err) != engine.ErrNoLogger {
		t.Errorf("err = (%v); want (engine.ErrNoLogger)", err)
	}
	if e != nil {
		t.Errorf("e = (%v); want (nil)", e)
	}
}

func TestEngineFailsNoCtx(t *testing.T) {
	t.Parallel()
	red := sampleReader(t, false)
	rec := sampleRecorder(t, false)

	e, err := engine.New(
		engine.WithRecorder(rec),
		engine.WithReaders(red),
		engine.WithLogger(tools.DiscardLogger()),
	)
	if errors.Cause(err) != engine.ErrNoCtx {
		t.Errorf("err = (%v); want (engine.ErrNoCtx)", err)
	}
	if e != nil {
		t.Errorf("e = (%v); want (nil)", e)
	}
}

func TestEventLoopOneReaderSendsPayload(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	red := sampleReader(t, true)
	rec := sampleRecorder(t, true)

	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithReaders(red),
		engine.WithLogger(log),
		engine.WithRecorder(rec),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
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

	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if job.ID != jobID {
			t.Errorf("job.ID = (%s); want (%s)", job.ID, jobID)
		}
		recorded <- struct{}{}
		return nil
	}

	done := make(chan struct{})
	go func() {
		engine.Start(e)
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
	t.Parallel()
	log := tools.DiscardLogger()
	log.Level = tools.DebugLevel
	ctx, cancel := context.WithCancel(context.Background())
	red1 := sampleReader(t, true)
	red2 := sampleReader(t, true)
	rec := sampleRecorder(t, true)
	red1.ReadFunc = func(job *token.Context) (*reader.Result, error) { return nil, nil }
	red2.ReadFunc = func(job *token.Context) (*reader.Result, error) { return nil, nil }
	rec.RecordFunc = func(context.Context, *recorder.Job) error { return nil }

	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithReaders(red1, red2),
		engine.WithLogger(log),
		engine.WithRecorder(rec),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}

	done := make(chan struct{})
	go func() {
		engine.Start(e)
		done <- struct{}{}
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}
}

func TestStartReadersTicking(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	rec := sampleRecorder(t, true)
	red := sampleReader(t, true)
	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithReaders(red),
		engine.WithLogger(log),
		engine.WithRecorder(rec),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}

	recorded := make(chan struct{})

	// Testing the engine ticks and sends a job request to the reader
	// There is no need for the actual job
	red.ReadFunc = func(*token.Context) (*reader.Result, error) {
		recorded <- struct{}{} // important, otherwise the test might not be valid
		return nil, errors.New("blah blah")
	}

	done := make(chan struct{})
	go func() {
		engine.Start(e)
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

func getReaderWithJob(t *testing.T, name string, jobContent []byte) *rdt.Reader {
	red, err := rdt.New(
		reader.WithLogger(tools.DiscardLogger()),
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

func eventLoopMultipleReadersSendPayloadEngine(t *testing.T, ctx context.Context) (*engine.Engine, *rct.Recorder, *rdt.Reader) {
	rec := sampleRecorder(t, false)
	red1 := getReaderWithJob(t, "reader1_name", []byte(`{"devil":666}`))
	red2 := getReaderWithJob(t, "reader2_name", []byte(`{"beelzebub":666}`))

	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithReaders(red1, red2),
		engine.WithLogger(log),
		engine.WithRecorder(rec),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	return e, rec, red1
}

// FIXME: break this test down
func TestEventLoopMultipleReadersSendPayload(t *testing.T) {
	t.Parallel()
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	e, rec, red1 := eventLoopMultipleReadersSendPayloadEngine(t, ctx)
	job1 := token.New(ctx)
	job2 := token.New(ctx)
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if job.ID != job1.ID() && job.ID != job2.ID() {
			t.Errorf("job.ID = (%s); want one of (%s, %s)", job.ID, job1.ID(), job2.ID())
		}
		return nil
	}

	wg.Add(1)
	go func() {
		engine.Start(e)
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
	case <-done2:
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

func TestWithReaderErrorWithEmptyReaderInput(t *testing.T) {
	t.Parallel()
	e, err := engine.New(
		engine.WithReaders(reader.DataReader(nil)),
	)
	if errors.Cause(err) != engine.ErrNoReader {
		t.Fatalf("err = (%#v); want (%v)", err, engine.ErrNoReader)
	}
	if e != nil {
		t.Fatalf("e = (%#v); want (nil)", e)
	}
}
