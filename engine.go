// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
	"context"
	"expvar"
	"fmt"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/config"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/recorder"
)

var (
	numGoroutines   = expvar.NewInt("Number Of Goroutines")
	expReaders      = expvar.NewInt("Readers")
	readJobs        = expvar.NewInt("Read Jobs")
	recordJobs      = expvar.NewInt("Record Jobs")
	erroredJobs     = expvar.NewInt("Error Jobs")
	contextCanceled = "context has been cancelled"
)

// Engine represents an engine that receives information from readers and ships them to a recorder.
// The Engine is allowed to change the index and type names at will.
// When the context times out or cancelled, the engine will close and return.
type Engine struct {
	log        logrus.FieldLogger
	ctx        context.Context            // Will call stop() when this context is cancelled/timed-out. This is a new context from the parent.
	name       string                     // Name identifier for this engine.
	recorder   recorder.DataRecorder      // Records to ElasticSearch client.
	readerJobs chan *reader.ReadJobResult // The results of reader jobs will be streamed here.

	wg      sync.WaitGroup
	redmu   sync.RWMutex
	readers []reader.DataReader // List of active readers.
}

// WithConfig creates an engine by instantiating readers and recorder from the configurations and sends them
// to the New function.
func WithConfig(ctx context.Context, log logrus.FieldLogger, recorderConf config.RecorderConf, readers ...config.ReaderConf) (*Engine, error) {

	reds := make([]reader.DataReader, len(readers))
	for i, redConf := range readers {
		red, err := redConf.NewInstance(ctx)
		if err != nil {
			return nil, err
		}
		reds[i] = red
	}

	rec, err := recorderConf.NewInstance(ctx)
	if err != nil {
		return nil, err
	}
	return New(ctx, log, rec, reds...)
}

// New creates an Engine instance with already set-up reader and recorders.
// The Engine's work starts from here by streaming all readers payloads to the recorder.
// Returns an error if there are recorders with the same name, or any of constructions results in errors.
func New(ctx context.Context, log logrus.FieldLogger, rec recorder.DataRecorder, reds ...reader.DataReader) (*Engine, error) {
	readerNames := make([]string, len(reds))
	seenNames := make(map[string]struct{}, len(reds))

	err := rec.Ping()
	if err != nil {
		return nil, ErrPing{Name: rec.Name(), Err: err}
	}

	for i, red := range reds {
		if _, ok := seenNames[red.Name()]; ok {
			return nil, ErrDuplicateRecorderName
		}
		err := red.Ping()
		if err != nil {
			return nil, ErrPing{Name: red.Name(), Err: err}
		}
		seenNames[red.Name()] = struct{}{}
		readerNames[i] = red.Name()
	}

	// just to be cute
	engineName := fmt.Sprintf("( %s >-x-<< %s )", rec.Name(), strings.Join(readerNames, ","))
	log = log.WithField("engine", engineName)
	cl := &Engine{
		name:       engineName,
		ctx:        ctx,
		readerJobs: make(chan *reader.ReadJobResult, len(reds)), // TODO: increase this is required
		recorder:   rec,
		readers:    reds,
		log:        log,
	}
	log.Debug("started the engine")
	return cl, nil
}

// setReaders is used in tests.
func (e *Engine) setReaders(readers []reader.DataReader) {
	e.redmu.Lock()
	defer e.redmu.Unlock()
	e.readers = readers
}
