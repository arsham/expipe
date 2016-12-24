// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package expvar contains logic to read from an expvar provide. The data comes in JSON format. The GC
// and memory related information will be changed to better presented to the data recorders.
// Bytes will be turned into megabyets, gc lists will be truncated to remove zero values.
package expvar

import (
	"context"
	"expvar"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/reader"
)

var (
	expvarReads = expvar.NewInt("Expvar Reads")
)

// Reader contains communication channels with a worker that exposes expvar information.
// It implements DataReader interface.
type Reader struct {
	name       string
	ctxReader  reader.ContextReader
	logger     logrus.FieldLogger
	mapper     datatype.Mapper
	typeName   string
	jobChan    chan context.Context
	resultChan chan *reader.ReadJobResult
	errorChan  chan<- communication.ErrorMessage
	interval   time.Duration
	timeout    time.Duration
}

// NewExpvarReader creates the worker and sets up its channels.
// Because the caller is reading the resp.Body, it is its job to close it.
func NewExpvarReader(
	logger logrus.FieldLogger,
	ctxReader reader.ContextReader,
	mapper datatype.Mapper,
	jobChan chan context.Context,
	resultChan chan *reader.ReadJobResult,
	errorChan chan<- communication.ErrorMessage,
	name string,
	typeName string,
	interval time.Duration,
	timeout time.Duration,
) (*Reader, error) {
	// TODO: ping the reader.
	// TODO: have the user decide how large the channels can be.
	logger = logger.WithField("reader", "expvar")
	w := &Reader{
		name:       name,
		typeName:   typeName,
		mapper:     mapper,
		jobChan:    jobChan,
		resultChan: resultChan,
		errorChan:  errorChan,
		ctxReader:  ctxReader,
		logger:     logger,
		timeout:    timeout,
		interval:   interval,
	}
	return w, nil
}

// Start begins reading from the target in its own goroutine.
// It will issue a goroutine on each job request.
// It will close the done channel when the job channel is closed.
func (r *Reader) Start(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	r.logger.Debug("starting")
	go func() {
	LOOP:
		for {
			select {
			case job := <-r.jobChan:
				go r.readMetrics(job)
			case <-ctx.Done():
				break LOOP
			}
		}
		close(done)
	}()
	return done
}

// Name shows the name identifier for this reader
func (r *Reader) Name() string { return r.name }

// TypeName shows the typeName the recorder should record as
func (r *Reader) TypeName() string { return r.typeName }

// Mapper returns the mapper object
func (r *Reader) Mapper() datatype.Mapper { return r.mapper }

// Interval returns the interval
func (r *Reader) Interval() time.Duration { return r.interval }

// Timeout returns the timeout
func (r *Reader) Timeout() time.Duration { return r.timeout }

// JobChan returns the job channel.
func (r *Reader) JobChan() chan context.Context { return r.jobChan }

// ResultChan returns the result channel.
func (r *Reader) ResultChan() chan *reader.ReadJobResult { return r.resultChan }

// will send an error back to the engine if it can't read from metrics provider
func (r *Reader) readMetrics(job context.Context) {
	resp, err := r.ctxReader.Get(job)
	id := communication.JobValue(job)
	if err != nil {
		r.logger.WithField("reader", "expvar_reader").
			WithField("name", r.Name()).
			WithField("ID", id).
			Debugf("%s: error making request: %v", r.name, err)
		r.errorChan <- communication.ErrorMessage{ID: id, Name: r.Name(), Err: err}
		return
	}

	res := &reader.ReadJobResult{
		ID:       id,
		Time:     time.Now(), // It is sensible to record the time now
		Res:      resp.Body,
		TypeName: r.TypeName(),
	}
	expvarReads.Add(1)
	r.resultChan <- res
}
