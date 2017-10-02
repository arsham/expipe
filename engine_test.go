// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expipe_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/arsham/expipe"
	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/internal/token"
	"github.com/arsham/expipe/reader"
	reader_testing "github.com/arsham/expipe/reader/testing"
	"github.com/arsham/expipe/recorder"
	recorder_testing "github.com/arsham/expipe/recorder/testing"

	"github.com/pkg/errors"
)

var (
	log        internal.FieldLogger
	testServer *httptest.Server
)

func TestMain(m *testing.M) {
	log = internal.DiscardLogger()
	testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	exitCode := m.Run()
	testServer.Close()
	os.Exit(exitCode)
}

func sampleReader() (*reader_testing.Reader, error) {
	return reader_testing.New(
		reader.SetLogger(log),
		reader.SetEndpoint(testServer.URL),
		reader.SetName("red_name"),
		reader.SetTypeName("type_name"),
		reader.SetInterval(time.Second),
		reader.SetTimeout(time.Second),
		reader.SetBackoff(5),
	)
}

func sampleRecorder() (*recorder_testing.Recorder, error) {
	return recorder_testing.New(
		recorder.SetLogger(log),
		recorder.SetEndpoint(testServer.URL),
		recorder.SetName("rec_name"),
		recorder.SetIndexName("index_name"),
		recorder.SetTimeout(time.Second),
		recorder.SetBackoff(5),
	)
}

func TestNewWithReadRecorder(t *testing.T) {
	ctx := context.Background()

	rec, err := sampleRecorder()
	if err != nil {
		t.Fatal(err)
	}
	red, err := sampleReader()
	if err != nil {
		t.Fatal(err)
	}
	red2, err := sampleReader()
	if err != nil {
		t.Fatal(err)
	}
	red2.SetName("d")

	e, err := expipe.EngineWithReadRecs(ctx, log, rec, map[string]reader.DataReader{red.Name(): red, red2.Name(): red2})
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if e == nil {
		t.Error("want Engine, got nil")
	}
}

func TestEngineSendJob(t *testing.T) {
	var recorderID token.ID
	ctx, cancel := context.WithCancel(context.Background())

	red, err := sampleReader()

	if err != nil {
		t.Fatal(err)
	}
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

	rec, err := sampleRecorder()
	if err != nil {
		t.Fatal(err)
	}
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if job.ID != recorderID {
			t.Errorf("want (%d), got (%s)", recorderID, job.ID)
		}
		return nil
	}

	e, err := expipe.EngineWithReadRecs(ctx, log, rec, map[string]reader.DataReader{red.Name(): red})
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	done := make(chan struct{})
	go func() {
		e.Start()
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

	rec, err := sampleRecorder()
	if err != nil {
		t.Fatal(err)
	}
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if !internal.StringInSlice(job.ID.String(), IDs) {
			t.Errorf("want once of (%s), got (%s)", strings.Join(IDs, ","), job.ID)
		}
		return nil
	}

	reds := make(map[string]reader.DataReader, count)
	for i := 0; i < count; i++ {

		name := fmt.Sprintf("reader_example_%d", i)
		red, err := sampleReader()
		if err != nil {
			t.Fatal(err)
		}
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
		reds[red.Name()] = red
	}

	e, err := expipe.EngineWithReadRecs(ctx, log, rec, reds)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	done := make(chan struct{})
	go func() {
		e.Start()
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
	ctx := context.Background()
	redServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	recServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer recServer.Close()
	redServer.Close() // making sure no one else is got this random port at this time

	rec, err := sampleRecorder()
	if err != nil {
		t.Fatal(err)
	}
	rec.SetEndpoint(recServer.URL)

	red, err := sampleReader()
	if err != nil {
		t.Fatal(err)
	}
	red.SetEndpoint(redServer.URL)

	e, err := expipe.EngineWithReadRecs(ctx, log, rec, map[string]reader.DataReader{red.Name(): red})
	if err == nil {
		t.Error("want ErrPing, got nil")
	}

	if _, ok := errors.Cause(err).(expipe.ErrPing); !ok {
		t.Errorf("want ErrPing, got (%v)", err)
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}
}

func TestEngineErrorsIfRecorderNotPinged(t *testing.T) {
	ctx := context.Background()
	redServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	recServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	recServer.Close() // making sure no one else is got this random port at this time
	defer redServer.Close()

	rec, err := sampleRecorder()
	if err != nil {
		t.Fatal(err)
	}
	rec.SetEndpoint(recServer.URL)

	red, err := sampleReader()
	if err != nil {
		t.Fatal(err)
	}

	e, err := expipe.EngineWithReadRecs(ctx, log, rec, map[string]reader.DataReader{red.Name(): red})
	if err == nil {
		t.Error("want ErrPing, got nil")
	}

	if _, ok := errors.Cause(err).(expipe.ErrPing); !ok {
		t.Errorf("want ErrPing, got (%v)", err)
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}
}

func TestEngineOnlyErrorsIfAllReadersNotPinged(t *testing.T) {
	ctx := context.Background()
	deadServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	liveServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer liveServer.Close()
	deadServer.Close() // making sure no one else is got this random port at this time

	rec, err := sampleRecorder()
	if err != nil {
		t.Fatal(err)
	}
	red1, err := sampleReader()
	if err != nil {
		t.Fatal(err)
	}
	red1.SetEndpoint(liveServer.URL)

	red2, err := sampleReader()
	if err != nil {
		t.Fatal(err)
	}
	red2.SetEndpoint(deadServer.URL)
	red2.SetName("b")
	red2.SetTypeName("ddb")

	e, err := expipe.EngineWithReadRecs(ctx, log, rec, map[string]reader.DataReader{red1.Name(): red1, red2.Name(): red2})
	if err != nil {
		t.Errorf("want nil, got (%v)", err)
	}

	if e == nil {
		t.Error("want Engine, got nil")
	}

	// now the engine should error
	red1, err = sampleReader()
	if err != nil {
		t.Fatal(err)
	}
	red1.SetEndpoint(deadServer.URL)
	red1.SetName("a")
	red1.SetTypeName("ddc")

	e, err = expipe.EngineWithReadRecs(ctx, log, rec, map[string]reader.DataReader{red1.Name(): red1, red2.Name(): red2})
	if err == nil {
		t.Error("want ErrPing, got nil")
	}

	if _, ok := errors.Cause(err).(expipe.ErrPing); !ok {
		t.Errorf("want ErrPing, got (%v)", err)
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}
}

func TestEngineShutsDownOnAllReadersGoOutOfScope(t *testing.T) {
	t.Parallel()
	stopReader1 := uint32(0)
	stopReader2 := uint32(0)
	readerInterval := time.Millisecond * 10

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	red1, err := sampleReader()
	if err != nil {
		t.Fatal(err)
	}
	red1.SetName("reader1_example")

	red2, err := sampleReader()
	if err != nil {
		t.Fatal(err)
	}
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

	rec, err := sampleRecorder()
	if err != nil {
		t.Fatal(err)
	}
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error { return nil }

	e, err := expipe.EngineWithReadRecs(ctx, log, rec, map[string]reader.DataReader{red1.Name(): red1, red2.Name(): red2})
	if err != nil {
		t.Fatal(err)
	}

	cleanExit := make(chan struct{})
	go func() {
		e.Start()
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

	red, err := sampleReader()
	if err != nil {
		t.Fatal(err)
	}
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

	rec, err := sampleRecorder()
	if err != nil {
		t.Fatal(err)
	}
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if atomic.LoadUint32(&stopRecorder) > 0 {
			return recorder.ErrBackoffExceeded
		}
		return nil
	}

	e, err := expipe.EngineWithReadRecs(ctx, log, rec, map[string]reader.DataReader{red.Name(): red})
	if err != nil {
		t.Fatal(err)
	}

	cleanExit := make(chan struct{})
	go func() {
		e.Start()
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
	ctx := context.Background()

	rec, err := sampleRecorder()
	if err != nil {
		t.Fatal(err)
	}

	e, err := expipe.New(
		expipe.SetCtx(ctx),
		expipe.SetLogger(log),
		expipe.SetRecorder(rec),
	)

	if errors.Cause(err) != expipe.ErrNoReader {
		t.Errorf("want ErrNoReader, got (%v)", err)
	}
	if e != nil {
		t.Errorf("want nil, got (%v)", e)
	}

	e, err = expipe.New(
		expipe.SetCtx(ctx),
		expipe.SetLogger(log),
		expipe.SetRecorder(rec),
	)
	if errors.Cause(err) != expipe.ErrNoReader {
		t.Errorf("want ErrNoReader, got (%v)", err)
	}
	if e != nil {
		t.Errorf("want nil, got (%v)", e)
	}
}

func TestFailsOnNilRecorder(t *testing.T) {
	e, err := expipe.New(
		expipe.SetRecorder(nil),
	)

	if err == nil {
		t.Error("want (error), got (nil)")
	}
	if e != nil {
		t.Errorf("want nil, got (%v)", e)
	}
}

func TestEngineFailsNoLog(t *testing.T) {
	red, err := sampleReader()
	if err != nil {
		t.Fatal(err)
	}
	rec, err := sampleRecorder()
	if err != nil {
		t.Fatal(err)
	}

	e, err := expipe.New(
		expipe.SetRecorder(rec),
		expipe.SetReaders(red),
		expipe.SetCtx(context.Background()),
	)
	if errors.Cause(err) != expipe.ErrNoLogger {
		t.Errorf("want (expipe.ErrNoLogger), got (%v)", err)
	}

	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}
}

func TestEngineFailsNoCtx(t *testing.T) {
	red, err := sampleReader()
	if err != nil {
		t.Fatal(err)
	}
	rec, err := sampleRecorder()
	if err != nil {
		t.Fatal(err)
	}

	e, err := expipe.New(
		expipe.SetRecorder(rec),
		expipe.SetReaders(red),
		expipe.SetLogger(internal.DiscardLogger()),
	)
	if errors.Cause(err) != expipe.ErrNoCtx {
		t.Errorf("want (expipe.ErrNoCtx), got (%v)", err)
	}

	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}
}
