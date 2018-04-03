// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine

import (
	"context"
	"expvar"
	"fmt"
	"strings"
	"sync"

	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/recorder"
	"github.com/arsham/expipe/tools"
	"github.com/pkg/errors"
)

var (
	numGoroutines     = expvar.NewInt("Number Of Goroutines")
	expReaders        = expvar.NewInt("Readers")
	readJobs          = expvar.NewInt("Read Jobs")
	waitingReadJobs   = expvar.NewInt("Waiting Read Jobs")
	recordJobs        = expvar.NewInt("Record Jobs")
	waitingRecordJobs = expvar.NewInt("Waiting Record Jobs")
	erroredJobs       = expvar.NewInt("Error Jobs")
	contextCanceled   = "context has been cancelled"
)

// Engine represents an engine that receives information from readers and ships
// them to a recorder. The Engine is allowed to change the index and type names
// at will. When the context times out or cancelled, the engine will close and
// return. Use the shutdown channel to signal the engine to stop recording.
// The ctx context will create a new context based on the parent.
type Engine struct {
	log        tools.FieldLogger
	ctx        context.Context       // Will call stop() when this context is cancelled/timed-out.
	name       string                // Name identifier for this engine.
	recorder   recorder.DataRecorder // Records to destination client.
	readerJobs chan *reader.Result   // The results of reader jobs.
	wg         sync.WaitGroup        // For keeping the reader counts.
	redmu      sync.RWMutex
	readers    map[string]reader.DataReader // Map of active readers name to their objects.
	shutdown   chan struct{}                // if closed, stops all operations and quits the engine.
}

func (e *Engine) String() string { return e.name }

// New generates the Engine based on the provided options.
func New(options ...func(*Engine) error) (*Engine, error) {
	e := &Engine{}
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
	if len(e.readers) == 0 {
		return nil, ErrNoReader
	}
	e.name = decorateName(e.readers, e.recorder)
	e.log = e.log.WithField("engine", e.name)
	return e, nil
}

func decorateName(readers map[string]reader.DataReader, recorder recorder.DataRecorder) string {
	var readerNames []string
	for name := range readers {
		readerNames = append(readerNames, name)
	}
	return fmt.Sprintf("( %s <-<< %s )", recorder.Name(), strings.Join(readerNames, ","))
}

// WithCtx uses ctx as the Engine's background context.
func WithCtx(ctx context.Context) func(*Engine) error {
	return func(e *Engine) error {
		e.ctx = ctx
		return nil
	}
}

// WithReaders builds up the readers and checks them.
func WithReaders(reds ...reader.DataReader) func(*Engine) error {
	return func(e *Engine) error {
		failedErrors := make(map[string]error)
		readers := make(map[string]reader.DataReader)
		for _, redConf := range reds {
			if redConf == nil {
				continue
			}
			err := redConf.Ping()
			if err != nil {
				failedErrors[redConf.Name()] = err
				continue
			}
			readers[redConf.Name()] = redConf
		}
		if len(failedErrors) > 0 && len(readers) == 0 {
			return PingError(failedErrors)
		}
		if len(readers) == 0 {
			return ErrNoReader
		}
		e.readers = readers
		e.readerJobs = make(chan *reader.Result, len(e.readers)) // TODO: increase this is as required (10)
		return nil
	}
}

// WithLogger sets the logger.
func WithLogger(log tools.FieldLogger) func(*Engine) error {
	return func(e *Engine) error {
		e.log = log
		return nil
	}
}

// WithRecorder builds up the recorder.
func WithRecorder(rec recorder.DataRecorder) func(*Engine) error {
	return func(e *Engine) error {
		if rec == nil {
			return errors.New("nil recorder")
		}
		err := rec.Ping()
		if err != nil {
			return PingError{rec.Name(): err}
		}
		e.recorder = rec
		return nil
	}
}
