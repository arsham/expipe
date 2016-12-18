// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"context"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
)

// ExpvarReader contains communication channels with a worker that exposes expvar information.
// It implements TargetReader interface.
type ExpvarReader struct {
	jobChan    chan context.Context
	resultChan chan *ReadJobResult
	ctxReader  ContextReader
	logger     logrus.FieldLogger
}

// NewExpvarReader creates the worker and sets up its channels.
// Because the caller is reading the resp.Body, it is its job to close it.
func NewExpvarReader(logger logrus.FieldLogger, ctxReader ContextReader) (*ExpvarReader, error) {
	// TODO: ping the reader.
	// TODO: have the user decide how large the channels can be.
	w := &ExpvarReader{
		jobChan:    make(chan context.Context, 1000),
		resultChan: make(chan *ReadJobResult, 1000),
		ctxReader:  ctxReader,
		logger:     logger,
	}
	return w, nil
}

// Start begins reading from the target in its own goroutine.
// It will issue a goroutine on each job request.
// It will close the done channel when the job channel is closed.
func (e *ExpvarReader) Start() chan struct{} {
	done := make(chan struct{})
	go func() {
		for job := range e.jobChan {
			go readMetrics(job, e.logger, e.ctxReader, e.resultChan)
		}
		close(done)
	}()
	return done
}

// JobChan returns the job channel.
func (e *ExpvarReader) JobChan() chan context.Context {
	return e.jobChan
}

// ResultChan returns the result channel.
func (e *ExpvarReader) ResultChan() chan *ReadJobResult {
	return e.resultChan
}

// will send an error back to the engine if it can't read from metrics provider
func readMetrics(job context.Context, logger logrus.FieldLogger, ctxReader ContextReader, resultChan chan *ReadJobResult) {
	resp, err := ctxReader.Get(job)
	if err != nil {
		logger.WithField("reader", "expvar_reader").Debugf("making request: %s", err)
		r := &ReadJobResult{
			Time: time.Now(),
			Res:  nil,
			Err:  fmt.Errorf("making request to metrics provider: %s", err),
		}
		resultChan <- r
		return
	}
	r := &ReadJobResult{
		Time: time.Now(), // It is sensible to record the time now
		Res:  resp.Body,
	}
	resultChan <- r
}
