// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
	"context"
	"time"

	"github.com/Sirupsen/logrus"
)

// ExpvarReader contains communication channels with a worker that exposes expvar information.
// It implements TargetReader interface
type ExpvarReader struct {
	jobChan    chan context.Context
	resultChan chan ReadJobResult
	ctxReader  ContextReader
	log        logrus.FieldLogger
}

// NewExpvarReader creates the worker and sets up its channels
// Because the caller is reading the resp.Body, it is its job to close it
func NewExpvarReader(log logrus.FieldLogger, ctxReader ContextReader) (*ExpvarReader, error) {
	// TODO: ping the reader
	w := &ExpvarReader{
		jobChan:    make(chan context.Context, 1000),
		resultChan: make(chan ReadJobResult, 1000),
		ctxReader:  ctxReader,
		log:        log,
	}
	return w, nil
}

// Start begins reading from the target in its own goroutine
// It will close the done channel when the job channel is closed
func (e *ExpvarReader) Start() chan struct{} {
	done := make(chan struct{})
	go func() {
		for job := range e.jobChan {
			// go goroutine
			r := ReadJobResult{}
			resp, err := e.ctxReader.ContextRead(job)
			if err != nil {
				e.log.Errorf("making request: %s", err)
				continue
			}
			r.Time = time.Now() // It is sensible to record the time now
			r.Res = resp.Body
			e.resultChan <- r
		}
		close(done)
	}()
	return done
}

// JobChan returns the job channel
func (e *ExpvarReader) JobChan() chan context.Context {
	return e.jobChan
}

// ResultChan returns the result channel
func (e *ExpvarReader) ResultChan() chan ReadJobResult {
	return e.resultChan
}
