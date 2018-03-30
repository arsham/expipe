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
	"github.com/arsham/expipe/reader"
	rdt "github.com/arsham/expipe/reader/testing"
	"github.com/arsham/expipe/recorder"
	rct "github.com/arsham/expipe/recorder/testing"
	"github.com/arsham/expipe/token"

	"github.com/pkg/errors"
)

var (
	log        internal.FieldLogger
	testServer *httptest.Server
)

func TestMain(m *testing.M) {
	log = internal.DiscardLogger()
	testServer = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)
	exitCode := m.Run()
	testServer.Close()
	os.Exit(exitCode)
}

func sampleReader() (*rdt.Reader, error) {
	return rdt.New(
		reader.WithLogger(log),
		reader.WithEndpoint(testServer.URL),
		reader.WithName("red_name"),
		reader.WithTypeName("type_name"),
		reader.WithInterval(time.Second),
		reader.WithTimeout(time.Second),
		reader.WithBackoff(5),
	)
}

func sampleRecorder() (*rct.Recorder, error) {
	return rct.New(
		recorder.WithLogger(log),
		recorder.WithEndpoint(testServer.URL),
		recorder.WithName("rec_name"),
		recorder.WithIndexName("index_name"),
		recorder.WithTimeout(time.Second),
		recorder.WithBackoff(5),
	)
}

func TestNewWithReadRecorder(t *testing.T) {
	ctx := context.Background()
	rec, err := sampleRecorder()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red, err := sampleReader()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red2, err := sampleReader()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red2.SetName("d")

	m := map[string]reader.DataReader{red.Name(): red, red2.Name(): red2}
	e, err := expipe.EngineWithReadRecs(ctx, log, rec, m)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if e == nil {
		t.Error("e = (nil); want (Engine)")
	}
}

func TestEngineSendJob(t *testing.T) {
	var recorderID token.ID
	ctx, cancel := context.WithCancel(context.Background())
	red, err := sampleReader()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
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
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if job.ID != recorderID {
			t.Errorf("want (%d), got (%s)", recorderID, job.ID)
		}
		return nil
	}

	m := map[string]reader.DataReader{red.Name(): red}
	e, err := expipe.EngineWithReadRecs(ctx, log, rec, m)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
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
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if !internal.StringInSlice(job.ID.String(), IDs) {
			t.Errorf("job.ID = (%s); want once of (%s)", job.ID, strings.Join(IDs, ","))
		}
		return nil
	}

	reds := make(map[string]reader.DataReader, count)
	for i := 0; i < count; i++ {
		var red *rdt.Reader
		name := fmt.Sprintf("reader_example_%d", i)
		red, err = sampleReader()
		if err != nil {
			t.Fatalf("err = (%#v); want (nil)", err)
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
		t.Errorf("err = (%v); want (nil)", err)
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
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	rec.SetEndpoint(recServer.URL)

	red, err := sampleReader()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red.SetEndpoint(redServer.URL)

	m := map[string]reader.DataReader{red.Name(): red}
	e, err := expipe.EngineWithReadRecs(ctx, log, rec, m)
	if err == nil {
		t.Error("err = (nil); want (PingError)")
	}

	if _, ok := errors.Cause(err).(expipe.PingError); !ok {
		t.Errorf("err = (%#v); want (PingError)", err)
	}
	if e != nil {
		t.Errorf("e = (%v); want (nil)", e)
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
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	rec.SetEndpoint(recServer.URL)

	red, err := sampleReader()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}

	m := map[string]reader.DataReader{red.Name(): red}
	e, err := expipe.EngineWithReadRecs(ctx, log, rec, m)
	if err == nil {
		t.Error("err = (nil); want (PingError)")
	}

	if _, ok := errors.Cause(err).(expipe.PingError); !ok {
		t.Errorf("err = (%v); want (PingError)", err)
	}
	if e != nil {
		t.Errorf("e = (%v); want (nil)", e)
	}
}

func TestEngineOnlyErrorsIfNoneOfReadersPinged(t *testing.T) {
	ctx := context.Background()
	deadServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	liveServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer liveServer.Close()
	deadServer.Close() // making sure no one else is got this random port at this time

	rec, err := sampleRecorder()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red1, err := sampleReader()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red1.SetEndpoint(liveServer.URL)

	red2, err := sampleReader()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red2.SetEndpoint(deadServer.URL)
	red2.SetName("b")
	red2.SetTypeName("ddb")

	m := map[string]reader.DataReader{red1.Name(): red1, red2.Name(): red2}
	e, err := expipe.EngineWithReadRecs(ctx, log, rec, m)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if e == nil {
		t.Error("e = (nil); want (Engine)")
	}

	// now the engine should error
	red1, err = sampleReader()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red1.SetEndpoint(deadServer.URL)
	red1.SetName("a")
	red1.SetTypeName("ddc")

	e, err = expipe.EngineWithReadRecs(ctx, log, rec, map[string]reader.DataReader{
		red1.Name(): red1,
		red2.Name(): red2,
	})
	if err == nil {
		t.Error("err = (nil); want (PingError)")
	}
	if _, ok := errors.Cause(err).(expipe.PingError); !ok {
		t.Errorf("err = (%v); want (PingError)", err)
	}
	if e != nil {
		t.Errorf("e = (%v); want (nil)", e)
	}
}

// FIXME: break this test down
func TestEngineShutsDownOnAllReadersGoOutOfScope(t *testing.T) {
	t.Parallel()
	stopReader1 := uint32(0)
	stopReader2 := uint32(0)
	readerInterval := time.Millisecond * 10

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	red1, err := sampleReader()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red1.SetName("reader1_example")

	red2, err := sampleReader()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
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
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error { return nil }

	m := map[string]reader.DataReader{red1.Name(): red1, red2.Name(): red2}
	e, err := expipe.EngineWithReadRecs(ctx, log, rec, m)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
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
		t.Fatalf("err = (%#v); want (nil)", err)
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
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if atomic.LoadUint32(&stopRecorder) > 0 {
			return recorder.ErrBackoffExceeded
		}
		return nil
	}

	m := map[string]reader.DataReader{red.Name(): red}
	e, err := expipe.EngineWithReadRecs(ctx, log, rec, m)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
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
		t.Fatalf("err = (%#v); want (nil)", err)
	}

	e, err := expipe.New(
		expipe.WithCtx(ctx),
		expipe.WithLogger(log),
		expipe.WithRecorder(rec),
	)
	if errors.Cause(err) != expipe.ErrNoReader {
		t.Errorf("err = (%v); want (NoReaderError)", err)
	}
	if e != nil {
		t.Errorf("e = (%v); want (Engine)", e)
	}

	e, err = expipe.New(
		expipe.WithCtx(ctx),
		expipe.WithLogger(log),
		expipe.WithRecorder(rec),
	)
	if errors.Cause(err) != expipe.ErrNoReader {
		t.Errorf("err = (%v); want (NoReaderError)", err)
	}
	if e != nil {
		t.Errorf("e = (%v); want (Engine)", e)
	}
}

func TestFailsOnNilRecorder(t *testing.T) {
	e, err := expipe.New(
		expipe.WithRecorder(nil),
	)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if e != nil {
		t.Errorf("e = (%v); want (Engine)", e)
	}
}

func TestEngineFailsNoLog(t *testing.T) {
	red, err := sampleReader()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	rec, err := sampleRecorder()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}

	e, err := expipe.New(
		expipe.WithRecorder(rec),
		expipe.WithReaders(red),
		expipe.WithCtx(context.Background()),
	)
	if errors.Cause(err) != expipe.ErrNoLogger {
		t.Errorf("err = (%v); want (expipe.ErrNoLogger)", err)
	}
	if e != nil {
		t.Errorf("e = (%v); want (nil)", e)
	}
}

func TestEngineFailsNoCtx(t *testing.T) {
	red, err := sampleReader()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	rec, err := sampleRecorder()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}

	e, err := expipe.New(
		expipe.WithRecorder(rec),
		expipe.WithReaders(red),
		expipe.WithLogger(internal.DiscardLogger()),
	)
	if errors.Cause(err) != expipe.ErrNoCtx {
		t.Errorf("err = (%v); want (expipe.ErrNoCtx)", err)
	}
	if e != nil {
		t.Errorf("e = (%v); want (nil)", e)
	}
}
