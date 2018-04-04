// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/arsham/expipe/datatype"

	"github.com/arsham/expipe/reader"
	rdt "github.com/arsham/expipe/reader/testing"
	"github.com/arsham/expipe/recorder"
	rct "github.com/arsham/expipe/recorder/testing"
	"github.com/arsham/expipe/tools/token"

	"github.com/arsham/expipe/engine"
	"github.com/arsham/expipe/tools"
	"github.com/pkg/errors"
)

var (
	log        tools.FieldLogger
	testServer *httptest.Server
	errExample = errors.New("error example")
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

func TestWithReaderNoReaderError(t *testing.T) {
	t.Parallel()
	e := &engine.Operator{}
	err := engine.WithReader(nil)(e)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
}

func TestWithReaderPinging(t *testing.T) {
	t.Parallel()
	e := &engine.Operator{}
	red := &rdt.Reader{
		PingFunc: func() error {
			return errExample
		},
	}
	err := engine.WithReader(red)(e)
	if _, ok := errors.Cause(err).(engine.PingError); !ok {
		t.Errorf("WithReader(): err = (%#v); want (engine.PingError)", err)
	}

	red.PingFunc = func() error {
		return nil
	}
	err = engine.WithReader(red)(e)
	if errors.Cause(err) != nil {
		t.Errorf("WithReader(): err = (%#v); want (nil)", err)
	}
}

func TestWithRecordersNoRecorderError(t *testing.T) {
	t.Parallel()
	e := &engine.Operator{}
	err := engine.WithRecorders()(e)
	if errors.Cause(err) != engine.ErrNoRecorder {
		t.Errorf("WithRecorders(): err = (%#v); want (engine.ErrNoRecorder)", err)
	}
}

func TestWithRecordersAtLeastOneRecorderPings(t *testing.T) {
	t.Parallel()
	e := &engine.Operator{}
	rec1 := &rct.Recorder{
		PingFunc: func() error { return nil },
	}
	rec2 := &rct.Recorder{
		PingFunc: func() error { return errExample },
	}
	err := engine.WithRecorders(rec1, rec2)(e)
	if errors.Cause(err) != nil {
		t.Errorf("WithRecorders(): err = (%#v); want (nil)", err)
	}
	err = engine.WithRecorders(rec1, nil)(e)
	if errors.Cause(err) != nil {
		t.Errorf("WithRecorders(): err = (%#v); want (nil)", err)
	}
}

func TestWithRecordersNoRecorderPingError(t *testing.T) {
	t.Parallel()
	e := &engine.Operator{}
	rec1 := &rct.Recorder{
		PingFunc: func() error { return errExample },
	}
	rec2 := &rct.Recorder{
		PingFunc: func() error { return errExample },
	}
	err := engine.WithRecorders(rec1, rec2)(e)
	if _, ok := errors.Cause(err).(engine.PingError); !ok {
		t.Errorf("WithRecorders(): err = (%#v); want (engine.PingError)", err)
	}
}

func TestNewNoLoggerError(t *testing.T) {
	t.Parallel()
	rec := &rct.Recorder{PingFunc: func() error { return nil }}
	red := &rdt.Reader{PingFunc: func() error { return nil }}
	e, err := engine.New(
		engine.WithCtx(context.Background()),
		engine.WithReader(red),
		engine.WithRecorders(rec),
	)
	if errors.Cause(err) == nil {
		t.Error("New(): err = (nil); want (error)")
	}
	if e != nil {
		t.Errorf("New(): e = (%#v); want (nil)", e)
	}
}

func TestNewNoCtxError(t *testing.T) {
	t.Parallel()
	rec := &rct.Recorder{PingFunc: func() error { return nil }}
	red := &rdt.Reader{PingFunc: func() error { return nil }}
	e, err := engine.New(
		engine.WithLogger(tools.DiscardLogger()),
		engine.WithReader(red),
		engine.WithRecorders(rec),
	)
	if errors.Cause(err) == nil {
		t.Error("New(): err = (nil); want (error)")
	}
	if e != nil {
		t.Errorf("New(): e = (%#v); want (nil)", e)
	}
}

func TestNewNoReaderError(t *testing.T) {
	t.Parallel()
	rec := &rct.Recorder{PingFunc: func() error { return nil }}
	e, err := engine.New(
		engine.WithCtx(context.Background()),
		engine.WithLogger(tools.DiscardLogger()),
		engine.WithRecorders(rec),
	)
	if errors.Cause(err) != engine.ErrNoReader {
		t.Errorf("New(): err = (%#v); want (engine.ErrNoReader)", err)
	}
	if e != nil {
		t.Errorf("New(): e = (%#v); want (nil)", e)
	}
}

func TestNewNoRecorderError(t *testing.T) {
	t.Parallel()
	red := &rdt.Reader{PingFunc: func() error { return nil }}
	e, err := engine.New(
		engine.WithCtx(context.Background()),
		engine.WithLogger(tools.DiscardLogger()),
		engine.WithReader(red),
	)
	if errors.Cause(err) != engine.ErrNoRecorder {
		t.Errorf("New(): err = (%#v); want (engine.ErrNoRecorder)", err)
	}
	if e != nil {
		t.Errorf("New(): e = (%#v); want (nil)", e)
	}
}

func TestNewOptionError(t *testing.T) {
	t.Parallel()
	red := &rdt.Reader{PingFunc: func() error { return errExample }}
	e, err := engine.New(
		engine.WithReader(red),
	)
	if errors.Cause(err) == nil {
		t.Errorf("New(): err = (%#v); want (error)", err)
	}
	if e != nil {
		t.Errorf("New(): e = (%#v); want (nil)", e)
	}
}

func TestNew(t *testing.T) {
	t.Parallel()
	red := &rdt.Reader{PingFunc: func() error { return nil }}
	rec := &rct.Recorder{PingFunc: func() error { return nil }}
	e, err := engine.New(
		engine.WithCtx(context.Background()),
		engine.WithLogger(tools.DiscardLogger()),
		engine.WithReader(red),
		engine.WithRecorders(rec),
	)
	if errors.Cause(err) != nil {
		t.Errorf("New(): err = (%#v); want (nil)", err)
	}
	if e == nil {
		t.Error("New(): e = (nil); want (Engine)")
	}
	if e.String() == "" {
		t.Error("String() (\"\"); want (Engine.String())")
	}
}

func TestSendJob(t *testing.T) {
	t.Parallel()
	recorderID := token.NewUID()
	received := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	red := &rdt.Reader{Pinged: true}
	red.ReadFunc = func(job *token.Context) (*reader.Result, error) {

		resp := &reader.Result{
			ID:       recorderID,
			Content:  []byte(`{"devil":666}`),
			TypeName: red.TypeName(),
			Mapper:   datatype.DefaultMapper(),
		}
		return resp, nil
	}

	f := func(ctx context.Context, job recorder.Job) error {
		if job.ID == recorderID {
			received <- struct{}{}
		}
		return nil
	}
	rec1 := &rct.Recorder{Pinged: true, MockName: "rec1"}
	rec2 := &rct.Recorder{Pinged: true, MockName: "rec2"}
	rec3 := &rct.Recorder{Pinged: true, MockName: "rec3"}
	rec4 := &rct.Recorder{Pinged: true, MockName: "rec4"}
	rec1.RecordFunc = f
	rec2.RecordFunc = f
	rec3.RecordFunc = f
	rec4.RecordFunc = f

	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithReader(red),
		engine.WithLogger(log),
		engine.WithRecorders(rec1, rec2, rec3, rec4),
	)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if e == nil {
		t.Fatal("e = (nil); want (Engine)")
	}

	done := engine.Start(e)

	for range e.Recorders() {
		select {
		case <-received:
		case <-time.After(time.Second):
			t.Error("didn't receive the job")
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}
}

// at interval, engine should ask the reader to send the payload
func TestTickingReader(t *testing.T) {
	t.Parallel()
	received := make(chan struct{})
	interval := 100 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	red := &rdt.Reader{
		Pinged:       true,
		MockInterval: interval,
	}
	red.ReadFunc = func(job *token.Context) (*reader.Result, error) {
		received <- struct{}{}
		return &reader.Result{}, nil
	}

	rec := &rct.Recorder{
		Pinged:   true,
		MockName: "rec1",
		RecordFunc: func(ctx context.Context, job recorder.Job) error {
			return nil
		},
	}

	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithReader(red),
		engine.WithLogger(log),
		engine.WithRecorders(rec),
	)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if e == nil {
		t.Fatal("e = (nil); want (Engine)")
	}

	done := engine.Start(e)

	// amount of ticks
	for i := 0; i < 3; i++ {
		select {
		case <-received:
		case <-time.After(time.Second * 2):
			t.Error("didn't receive the job order")
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}

}
