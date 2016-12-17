// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
    "context"
    "fmt"
    "io"
    "time"

    "github.com/Sirupsen/logrus"
    "github.com/arsham/expvastic/lib"
)

// ExpvarReader contains communication channels with the worker
type ExpvarReader struct {
    jobCh     chan context.Context
    resultCh  chan JobResult
    ctxReader ContextReader
}

// Job is sent with a context and a channel to read the errors
type Job struct {
    Payload context.Context
    Err     chan error
}

// JobResult is constructed everytime a new record is fetched.
// The time is set after the request was successfully read
type JobResult struct {
    Time time.Time
    Res  io.ReadCloser
    Err  error
}

// NewExpvarReader creates the worker and sets up its channels
// Because the caller is reading the resp.Body, it is its job to close it
func NewExpvarReader(log logrus.FieldLogger, ctxReader ContextReader) (*ExpvarReader, error) {
    // TODO: ping the reader
    w := &ExpvarReader{
        jobCh:     make(chan context.Context, 1000),
        resultCh:  make(chan JobResult, 1000),
        ctxReader: ctxReader,
    }
    return w, nil
}

// Start begins reading from the target in its own goroutine
// It will close the done channel when the job channel is closed
func (e *ExpvarReader) Start() chan struct{} {
    done := make(chan struct{})
    go func() {
        for job := range e.jobCh {
            r := JobResult{}
            resp, err := e.ctxReader.ContextRead(job)
            if err != nil {
                r.Err = fmt.Errorf("making request: %s", err)
                r.Res = new(lib.DummyReadCloser)
                e.resultCh <- r
                continue
            }
            r.Time = time.Now() // It is sensible to record the time now
            r.Res = resp.Body
            e.resultCh <- r
        }
        close(done)
    }()
    return done
}

// JobChan returns the job channel
func (e *ExpvarReader) JobChan() chan context.Context {
    return e.jobCh
}

// ResultChan returns the result channel
func (e *ExpvarReader) ResultChan() chan JobResult {
    return e.resultCh
}
