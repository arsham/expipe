// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/arsham/expipe/tools"

	"github.com/arsham/expipe/engine"
	"github.com/arsham/expipe/reader"
	rdt "github.com/arsham/expipe/reader/testing"
	"github.com/arsham/expipe/recorder"
	rct "github.com/arsham/expipe/recorder/testing"
	"github.com/arsham/expipe/tools/config"
	"github.com/arsham/expipe/tools/token"
	"github.com/pkg/errors"
)

func TestStartEmptyConfmapError(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := newFakeLogger()

	s := &engine.Service{Log: log, Ctx: ctx}
	done, err := s.Start()
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if done != nil {
		t.Errorf("d = (%#v); want (nil)", done)
	}
}

func TestStartEmptyReadersError(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := newFakeLogger()
	confMap := &config.ConfMap{
		Readers:   map[string]reader.DataReader{},
		Recorders: map[string]recorder.DataRecorder{"rec1": &rct.Recorder{}},
		Routes:    map[string][]string{"red1": {"rec1", "rec2"}},
	}
	s := &engine.Service{Log: log, Ctx: ctx, Conf: confMap}
	done, err := s.Start()
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if done != nil {
		t.Errorf("d = (%v); want (nil)", done)
	}
	confMap = &config.ConfMap{
		Readers:   map[string]reader.DataReader{"red1": nil},
		Recorders: map[string]recorder.DataRecorder{"rec1": &rct.Recorder{}},
		Routes:    map[string][]string{"red1": {"rec1", "rec2"}},
	}

	s = &engine.Service{Log: log, Ctx: ctx, Conf: confMap}
	done, err = s.Start()
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if done != nil {
		t.Errorf("d = (%v); want (nil)", done)
	}
}

func TestStartEmptyRecordersErrors(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := newFakeLogger()
	confMap := &config.ConfMap{
		Readers:   map[string]reader.DataReader{"red1": &rdt.Reader{}},
		Recorders: map[string]recorder.DataRecorder{},
		Routes:    map[string][]string{"red1": {"rec1", "rec2"}},
	}

	s := &engine.Service{Log: log, Ctx: ctx, Conf: confMap}
	done, err := s.Start()
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if done != nil {
		t.Errorf("d = (%v); want (nil)", done)
	}
	confMap = &config.ConfMap{
		Readers:   map[string]reader.DataReader{"red1": &rdt.Reader{}},
		Recorders: map[string]recorder.DataRecorder{"rec1": nil},
		Routes:    map[string][]string{"red1": {"rec1", "rec2"}},
	}

	s = &engine.Service{Log: log, Ctx: ctx, Conf: confMap}
	done, err = s.Start()
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if done != nil {
		t.Errorf("d = (%v); want (nil)", done)
	}
}

func TestStartEmptyReaderNameErrors(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := newFakeLogger()
	confMap := &config.ConfMap{
		Readers:   map[string]reader.DataReader{"red1": &rdt.Reader{}},
		Recorders: map[string]recorder.DataRecorder{"rec1": &rct.Recorder{MockName: "name"}},
		Routes:    map[string][]string{"red1": {"rec1", "rec2"}},
	}

	s := &engine.Service{Log: log, Ctx: ctx, Conf: confMap}
	done, err := s.Start()
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if done != nil {
		t.Errorf("d = (%v); want (nil)", done)
	}
}

func TestStartEmptyRcorderNameErrors(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := newFakeLogger()
	confMap := &config.ConfMap{
		Readers:   map[string]reader.DataReader{"red1": &rdt.Reader{MockName: "name"}},
		Recorders: map[string]recorder.DataRecorder{"rec1": &rct.Recorder{}},
		Routes:    map[string][]string{"red1": {"rec1", "rec2"}},
	}

	s := &engine.Service{Log: log, Ctx: ctx, Conf: confMap}
	done, err := s.Start()
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if done != nil {
		t.Errorf("d = (%v); want (nil)", done)
	}
}

func TestStartRecorderNotFoundErrors(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := newFakeLogger()
	confMap := &config.ConfMap{
		Readers:   map[string]reader.DataReader{"red1": &rdt.Reader{MockName: "name"}},
		Recorders: map[string]recorder.DataRecorder{"rec1": &rct.Recorder{MockName: "name"}},
		Routes:    map[string][]string{"red1": {"rec2", "rec3"}},
	}

	s := &engine.Service{Log: log, Ctx: ctx, Conf: confMap}
	done, err := s.Start()
	if errors.Cause(err) != engine.ErrNoRecorder {
		t.Error("err = (nil); want (engine.ErrNoRecorder)")
	}
	if done != nil {
		t.Errorf("d = (%v); want (nil)", done)
	}
}

type operator struct {
	engine.Engine
	red  reader.DataReader
	recs map[string]recorder.DataRecorder
	log  tools.FieldLogger
	ctx  context.Context
}

func (o *operator) Reader() reader.DataReader                   { return o.red }
func (o *operator) Recorders() map[string]recorder.DataRecorder { return o.recs }
func (o *operator) Ctx() context.Context                        { return o.ctx }
func (o *operator) Log() tools.FieldLogger                      { return o.log }

func TestStartCallsStart(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	called := make(chan struct{})
	log := newFakeLogger()
	red := &rdt.Reader{MockName: "name"}
	rec := &rct.Recorder{MockName: "name"}
	confMap := &config.ConfMap{
		Readers:   map[string]reader.DataReader{"red": red},
		Recorders: map[string]recorder.DataRecorder{"rec": rec},
		Routes:    map[string][]string{"red": {"rec"}},
	}
	o := &operator{
		log: log, ctx: ctx, red: red,
		recs: map[string]recorder.DataRecorder{
			rec.MockName: rec,
		},
	}
	s := &engine.Service{
		Log: log, Ctx: ctx, Conf: confMap,
		Configure: func(...func(engine.Engine) error) (engine.Engine, error) {
			close(called)
			return o, nil
		},
	}
	done, err := s.Start()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	if done == nil {
		t.Error("d = (nil); want (Engine)")
	}
	select {
	case <-called:
	case <-time.After(time.Second):
		t.Error("Configure wasn't called")
	}
}

func TestStartSendsReadRequestToReader(t *testing.T) {
	if testing.Short() {
		return
	}
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	called := make(chan struct{})
	log := newFakeLogger()
	log.ErrorfFunc = func(string, ...interface{}) {}

	red := &rdt.Reader{
		MockName: "name",
		ReadFunc: func(*token.Context) (*reader.Result, error) {
			called <- struct{}{}
			return nil, nil
		},
		MockInterval: time.Second,
	}
	rec := &rct.Recorder{MockName: "name"}
	confMap := &config.ConfMap{
		Readers:   map[string]reader.DataReader{"red": red},
		Recorders: map[string]recorder.DataRecorder{"rec": rec},
		Routes:    map[string][]string{"red": {"rec"}},
	}
	o := &operator{
		ctx: ctx, log: log, red: red,
		recs: map[string]recorder.DataRecorder{
			rec.MockName: rec,
		},
	}
	s := &engine.Service{
		Ctx: ctx, Log: log, Conf: confMap,
		Configure: func(...func(engine.Engine) error) (engine.Engine, error) {
			return o, nil
		},
	}
	s.Start()
	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Error("Configure wasn't called")
	}
}

func TestStartFinishesWhenContextIsCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := newFakeLogger()
	log.ErrorfFunc = func(string, ...interface{}) {}

	red := &rdt.Reader{MockName: "name"}
	rec := &rct.Recorder{MockName: "name"}
	confMap := &config.ConfMap{
		Readers:   map[string]reader.DataReader{"red": red},
		Recorders: map[string]recorder.DataRecorder{"rec": rec},
		Routes:    map[string][]string{"red": {"rec"}},
	}
	o := &operator{
		ctx: ctx, log: log, red: red,
		recs: map[string]recorder.DataRecorder{
			rec.MockName: rec,
		},
	}
	s := &engine.Service{
		Ctx: ctx, Log: log, Conf: confMap,
		Configure: func(...func(engine.Engine) error) (engine.Engine, error) {
			return o, nil
		},
	}
	done, err := s.Start()
	if err != nil {
		t.Fatalf("Start(): err = (%#v); want (nil)", err)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("Service didn't quit")
	}
}
