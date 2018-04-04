// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine_test

import (
	"bytes"
	"context"
	"strings"
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

type fakeLogger struct {
	tools.FieldLogger
	ErrorFunc  func(args ...interface{})
	ErrorfFunc func(format string, args ...interface{})
}

func newFakeLogger() *fakeLogger {
	return &fakeLogger{
		FieldLogger: tools.DiscardLogger(),
		ErrorFunc:   func(args ...interface{}) {},
		ErrorfFunc:  func(format string, args ...interface{}) {},
	}
}
func (f fakeLogger) Error(args ...interface{})                 { f.ErrorFunc(args...) }
func (f fakeLogger) Errorf(format string, args ...interface{}) { f.ErrorfFunc(format, args...) }

func TestStartStopsOnCanceledContext(t *testing.T) {
	t.Parallel()
	log := newFakeLogger()
	ctx, cancel := context.WithCancel(context.Background())
	red := &rdt.Reader{
		PingFunc:     func() error { return nil },
		MockInterval: time.Second,
		MockMapper:   datatype.DefaultMapper(),
	}
	rec := &rct.Recorder{
		PingFunc: func() error { return nil },
	}
	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithLogger(log),
		engine.WithReader(red),
		engine.WithRecorders(rec),
	)
	if errors.Cause(err) != nil {
		t.Errorf("New(): err = (%#v); want (nil)", err)
	}
	if e == nil {
		t.Fatal("New(): e = (nil); want (Engine)")
	}
	stop := engine.Start(e)
	if stop == nil {
		t.Error("Star() = (nil); want (chan struct{})")
	}
	select {
	case <-stop:
		t.Error("stop was closed")
	case <-time.After(time.Millisecond * 100):
	}
	cancel()
	select {
	case <-stop:
	case <-time.After(time.Millisecond * 100):
		t.Error("stop didn't closed")
	}
}

func TestStartInvokesRead(t *testing.T) {
	t.Parallel()
	log := newFakeLogger()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	called := make(chan struct{})
	interval := time.Second
	red := &rdt.Reader{
		PingFunc: func() error { return nil },
		ReadFunc: func(*token.Context) (*reader.Result, error) {
			called <- struct{}{}
			return nil, nil
		},
		MockInterval: interval,
		MockMapper:   datatype.DefaultMapper(),
	}
	rec := &rct.Recorder{
		PingFunc: func() error { return nil },
	}
	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithLogger(log),
		engine.WithReader(red),
		engine.WithRecorders(rec),
	)
	if errors.Cause(err) != nil {
		t.Errorf("New(): err = (%#v); want (nil)", err)
	}
	if e == nil {
		t.Fatal("New(): e = (nil); want (Engine)")
	}

	engine.Start(e)
	select {
	case <-called:
	case <-time.After(interval * 1000):
		t.Error("didn't invoke the read")
	}
}

func TestReadError(t *testing.T) {
	t.Parallel()
	log := newFakeLogger()
	called := make(chan struct{})
	log.ErrorfFunc = func(string, ...interface{}) {
		called <- struct{}{}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	interval := time.Millisecond
	red := &rdt.Reader{
		PingFunc: func() error { return nil },
		ReadFunc: func(*token.Context) (*reader.Result, error) {
			return nil, errExample
		},
		MockInterval: interval,
		MockMapper:   datatype.DefaultMapper(),
	}
	rec := &rct.Recorder{
		PingFunc: func() error { return nil },
	}
	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithLogger(log),
		engine.WithReader(red),
		engine.WithRecorders(rec),
	)
	engine.WithLogger(log)(e)
	if errors.Cause(err) != nil {
		t.Errorf("New(): err = (%#v); want (nil)", err)
	}
	if e == nil {
		t.Fatal("New(): e = (nil); want (Engine)")
	}

	engine.Start(e)
	select {
	case <-called:
	case <-time.After(interval * 1000):
		t.Error("didn't register the error")
	}
}

func TestReadErrorOnNilJob(t *testing.T) {
	t.Parallel()
	log := newFakeLogger()
	called := make(chan struct{})
	log.ErrorfFunc = func(string, ...interface{}) {
		called <- struct{}{}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	interval := time.Millisecond
	red := &rdt.Reader{
		PingFunc: func() error { return nil },
		ReadFunc: func(*token.Context) (*reader.Result, error) {
			return nil, nil
		},
		MockInterval: interval,
		MockMapper:   datatype.DefaultMapper(),
	}
	rec := &rct.Recorder{
		PingFunc: func() error { return nil },
	}
	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithLogger(log),
		engine.WithReader(red),
		engine.WithRecorders(rec),
	)
	engine.WithLogger(log)(e)
	if errors.Cause(err) != nil {
		t.Errorf("New(): err = (%#v); want (nil)", err)
	}
	if e == nil {
		t.Fatal("New(): e = (nil); want (Engine)")
	}

	engine.Start(e)
	select {
	case <-called:
	case <-time.After(interval * 1000):
		t.Error("didn't register the error")
	}
}

func TestReadErrorBadPayload(t *testing.T) {
	t.Parallel()
	log := newFakeLogger()
	registered := make(chan struct{})
	recorded := make(chan struct{})
	log.ErrorfFunc = func(string, ...interface{}) {
		close(registered)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	interval := time.Millisecond
	job := token.New(ctx)
	jobID := job.ID()
	red := &rdt.Reader{
		PingFunc:     func() error { return nil },
		MockInterval: interval,
		MockMapper:   datatype.DefaultMapper(),
	}
	red.ReadFunc = func(*token.Context) (*reader.Result, error) {
		resp := &reader.Result{
			ID:       jobID,
			Content:  []byte(`{"god":777`),
			TypeName: red.TypeName(),
			Mapper:   red.Mapper(),
		}
		return resp, nil

	}
	rec := &rct.Recorder{
		PingFunc: func() error { return nil },
		RecordFunc: func(context.Context, recorder.Job) error {
			close(recorded)
			return nil
		},
	}
	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithLogger(log),
		engine.WithReader(red),
		engine.WithRecorders(rec),
	)
	engine.WithLogger(log)(e)
	if errors.Cause(err) != nil {
		t.Errorf("New(): err = (%#v); want (nil)", err)
	}
	if e == nil {
		t.Fatal("New(): e = (nil); want (Engine)")
	}

	engine.Start(e)
	select {
	case <-registered:
	case <-time.After(interval * 1000):
		t.Error("didn't register the error")
	}
	select {
	case <-recorded:
		t.Error("shouldn't have recorded")
	case <-time.After(interval * 10):
	}
}

func TestRecordError(t *testing.T) {
	t.Parallel()
	log := newFakeLogger()
	registered := make(chan struct{})
	log.ErrorfFunc = func(string, ...interface{}) {
		registered <- struct{}{}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	interval := time.Millisecond
	job := token.New(ctx)
	jobID := job.ID()
	red := &rdt.Reader{
		PingFunc:     func() error { return nil },
		MockInterval: interval,
		MockMapper:   datatype.DefaultMapper(),
	}
	red.ReadFunc = func(*token.Context) (*reader.Result, error) {
		resp := &reader.Result{
			ID:       jobID,
			Content:  []byte(`{"lucifer":666}`),
			TypeName: red.TypeName(),
			Mapper:   red.Mapper(),
		}
		return resp, nil

	}
	rec := &rct.Recorder{
		PingFunc: func() error { return nil },
		RecordFunc: func(context.Context, recorder.Job) error {
			return errExample
		},
	}
	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithLogger(log),
		engine.WithReader(red),
		engine.WithRecorders(rec),
	)
	engine.WithLogger(log)(e)
	if errors.Cause(err) != nil {
		t.Errorf("New(): err = (%#v); want (nil)", err)
	}
	if e == nil {
		t.Fatal("New(): e = (nil); want (Engine)")
	}

	engine.Start(e)
	select {
	case <-registered:
	case <-time.After(interval * 1000):
		t.Error("didn't register the error")
	}
}

func TestEngineDispatchesToRecorders(t *testing.T) {
	t.Parallel()
	log := newFakeLogger()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recorded := make(chan struct{})
	interval := time.Millisecond
	job := token.New(ctx)
	jobID := job.ID()
	content := []byte(`{"devil":666}`)
	partial := `"devil":666`
	now := time.Now()
	red := &rdt.Reader{
		PingFunc:     func() error { return nil },
		MockInterval: interval,
		MockMapper:   datatype.DefaultMapper(),
	}
	red.ReadFunc = func(*token.Context) (*reader.Result, error) {
		resp := &reader.Result{
			ID:       jobID,
			Content:  content,
			TypeName: red.TypeName(),
			Mapper:   red.Mapper(),
		}
		return resp, nil
	}

	recordFunc := func(ctx context.Context, job recorder.Job) error {
		if job.ID != jobID {
			t.Errorf("job.ID = (%s); want (%s)", job.ID, jobID)
		}
		p := new(bytes.Buffer)
		job.Payload.Generate(p, now)
		if !strings.Contains(p.String(), partial) {
			t.Errorf("content = (%s); want (%s)", p.Bytes(), content)
		}
		recorded <- struct{}{}
		return nil
	}
	rec1 := &rct.Recorder{
		MockName:   "rec1",
		PingFunc:   func() error { return nil },
		RecordFunc: recordFunc,
	}
	rec2 := &rct.Recorder{
		MockName:   "rec2",
		PingFunc:   func() error { return nil },
		RecordFunc: recordFunc,
	}
	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithLogger(log),
		engine.WithReader(red),
		engine.WithRecorders(rec1, rec2),
	)
	engine.WithLogger(log)(e)
	if errors.Cause(err) != nil {
		t.Errorf("New(): err = (%#v); want (nil)", err)
	}
	if e == nil {
		t.Fatal("New(): e = (nil); want (Engine)")
	}

	engine.Start(e)
	select {
	case <-recorded:
	case <-time.After(interval * 1000):
		t.Error("expected to record, didn't happen")
	}
	select {
	case <-recorded:
	case <-time.After(interval * 1000):
		t.Error("expected to record, didn't happen")
	}
}
