// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine

import (
	"context"
	"expvar"
	"fmt"
	"strings"

	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/recorder"
	"github.com/arsham/expipe/tools"
	"github.com/pkg/errors"
)

var (
	numGoroutines     = expvar.NewInt("Number Of Goroutines")
	expRecorders      = expvar.NewInt("Recorders")
	readJobs          = expvar.NewInt("Read Jobs")
	waitingReadJobs   = expvar.NewInt("Waiting Read Jobs")
	recordJobs        = expvar.NewInt("Record Jobs")
	waitingRecordJobs = expvar.NewInt("Waiting Record Jobs")
	erroredJobs       = expvar.NewInt("Error Jobs")
)

// Engine is an interface to Operator's behaviour.
// This abstraction is very tight on purpose.
type Engine interface {
	fmt.Stringer
	SetCtx(context.Context)
	SetLog(tools.FieldLogger)
	SetRecorders(map[string]recorder.DataRecorder)
	SetReader(reader.DataReader)
	Ctx() context.Context
	Log() tools.FieldLogger
	Recorders() map[string]recorder.DataRecorder
	Reader() reader.DataReader
}

// Operator represents an Engine that receives information from a reader and
// ships them to multiple recorders.
type Operator struct {
	log       tools.FieldLogger
	ctx       context.Context // Will call stop() when this context is cancelled/timed-out.
	name      string          // Name identifier for this Engine.
	reader    reader.DataReader
	recorders map[string]recorder.DataRecorder // Map of active recorders name to their objects.
}

func (o *Operator) String() string { return o.name }

// Ctx returns the context assigned to this Engine.
func (o Operator) Ctx() context.Context { return o.ctx }

// Log returns the logger assigned to this Engine.
func (o Operator) Log() tools.FieldLogger { return o.log }

// Recorders returns the recorder map.
func (o Operator) Recorders() map[string]recorder.DataRecorder { return o.recorders }

// Reader returns the reader.
func (o Operator) Reader() reader.DataReader { return o.reader }

// SetCtx sets the context of this Engine.
func (o *Operator) SetCtx(ctx context.Context) { o.ctx = ctx }

// SetLog sets the logger of this Engine.
func (o *Operator) SetLog(log tools.FieldLogger) { o.log = log }

// SetRecorders sets the recorder map.
func (o *Operator) SetRecorders(recorders map[string]recorder.DataRecorder) {
	o.recorders = recorders
}

// SetReader sets the reader.
func (o *Operator) SetReader(reader reader.DataReader) { o.reader = reader }

// New generates the Engine based on the provided options.
func New(options ...func(Engine) error) (Engine, error) {
	e := &Operator{}
	for _, op := range options {
		err := op(e)
		if err != nil {
			return nil, errors.Wrap(err, "option creation")
		}
	}
	if e.log == nil {
		return nil, ErrNoLogger
	}
	if e.ctx == nil {
		return nil, ErrNoCtx
	}
	if len(e.recorders) == 0 {
		return nil, ErrNoRecorder
	}
	if e.reader == nil {
		return nil, ErrNoReader
	}
	e.name = decorateName(e.reader, e.recorders)
	e.log = e.log.WithField("engine", e.name)
	return e, nil
}

func decorateName(reader reader.DataReader, recorders map[string]recorder.DataRecorder) string {
	var recNames []string
	for name := range recorders {
		recNames = append(recNames, name)
	}
	return fmt.Sprintf("( %s >->> %s )", reader.Name(), strings.Join(recNames, ","))
}

// WithCtx uses ctx as the Engine's background context.
func WithCtx(ctx context.Context) func(Engine) error {
	return func(e Engine) error {
		e.SetCtx(ctx)
		return nil
	}
}

// WithRecorders builds up the recorder and checks them.
func WithRecorders(recs ...recorder.DataRecorder) func(Engine) error {
	return func(e Engine) error {
		failedErrors := make(map[string]error)
		recorders := make(map[string]recorder.DataRecorder)
		for _, rec := range recs {
			if rec == nil {
				continue
			}
			err := rec.Ping()
			if err != nil {
				failedErrors[rec.Name()] = err
				continue
			}
			recorders[rec.Name()] = rec
			expRecorders.Add(1)
		}
		if len(failedErrors) > 0 && len(recorders) == 0 {
			return PingError(failedErrors)
		}
		if len(recorders) == 0 {
			return ErrNoRecorder
		}
		e.SetRecorders(recorders)
		return nil
	}
}

// WithLogger sets the logger.
func WithLogger(log tools.FieldLogger) func(Engine) error {
	return func(e Engine) error {
		e.SetLog(log)
		return nil
	}
}

// WithReader builds up the reader.
func WithReader(red reader.DataReader) func(Engine) error {
	return func(e Engine) error {
		if red == nil {
			return errors.New("nil reader")
		}
		err := red.Ping()
		if err != nil {
			return PingError{red.Name(): err}
		}
		e.SetReader(red)
		return nil
	}
}
