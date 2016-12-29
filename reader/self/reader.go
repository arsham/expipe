// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package self contains codes for recording expvastic's own metrics.
// Here is a list of currently collected metrics:
//
//      ElasticSearch Var Name    | expvastic var name
//      ----------------------------------------------
//      Recorders                 | expRecorders
//      Readers                   | expReaders
//      Read Jobs                 | readJobs
//      Record Jobs               | recordJobs
//      Errored Jobs              | erroredJobs
//      Records Distributed       | recordsDistributed
//      DataType Objects          | datatypeObjs
//      DataType Objects Errors   | datatypeErrs
//      Unidentified JSON Count   | unidentifiedJSON
//      StringType Count          | stringTypeCount
//      FloatType Count           | floatTypeCount
//      GCListType Count          | gcListTypeCount
//      ByteType Count            | byteTypeCount
//      Expvar Reads              | expvarReads
//      ElasticSearch Records     | elasticsearchRecords
package self

import (
	"context"
	// to expose the metrics
	_ "expvar"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
)

// Reader contains communication channels with a worker that exposes expvar information.
// It implements DataReader interface.
type Reader struct {
	name       string
	typeName   string
	log        logrus.FieldLogger
	mapper     datatype.Mapper
	jobChan    chan context.Context
	resultChan chan *reader.ReadJobResult
	errorChan  chan<- communication.ErrorMessage
	interval   time.Duration
	timeout    time.Duration
	url        string
}

// NewSelfReader exposes expvastic's own metrics.
func NewSelfReader(
	log logrus.FieldLogger,
	endpoint string,
	mapper datatype.Mapper,
	jobChan chan context.Context,
	resultChan chan *reader.ReadJobResult,
	errorChan chan<- communication.ErrorMessage,
	name,
	typeName string,
	interval time.Duration,
	timeout time.Duration,
) (*Reader, error) {
	if name == "" {
		return nil, reader.ErrEmptyName
	}

	if endpoint == "" {
		return nil, reader.ErrEmptyEndpoint
	}

	url, err := lib.SanitiseURL(endpoint)
	if err != nil {
		return nil, reader.ErrInvalidEndpoint(endpoint)
	}
	_, err = http.Head(url)
	if err != nil {
		return nil, reader.ErrEndpointNotAvailable{Endpoint: url, Err: err}
	}

	if typeName == "" {
		return nil, reader.ErrEmptyTypeName
	}

	log = log.WithField("engine", "expvastic")
	w := &Reader{
		name:       name,
		typeName:   typeName,
		mapper:     mapper,
		jobChan:    jobChan,
		resultChan: resultChan,
		errorChan:  errorChan,
		log:        log,
		interval:   interval,
		timeout:    timeout,
		url:        url,
	}
	return w, nil
}

// Start begins reading from the target in its own goroutine.
// It will issue a goroutine on each job request.
// It will close the done channel when the job channel is closed.
func (r *Reader) Start(ctx context.Context, stop communication.StopChannel) {
	r.log.Debug("starting")
	go func() {
		for {
			select {
			case job := <-r.jobChan:
				go r.readMetrics(job)
			case s := <-stop:
				s <- struct{}{}
				return
			}
		}
	}()
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
	id := communication.JobValue(job)
	resp, err := http.Get(r.url)
	if err != nil {
		r.log.WithField("reader", "self").
			WithField("ID", id).
			Errorf("%s: error making request: %v", r.name, err) // Error because it is a local dependency.
		r.errorChan <- communication.ErrorMessage{ID: id, Name: r.Name(), Err: err}
		return
	}

	res := &reader.ReadJobResult{
		ID:       id,
		Time:     time.Now(), // It is sensible to record the time now
		Res:      resp.Body,
		TypeName: r.TypeName(),
		Mapper:   r.Mapper(),
	}
	r.resultChan <- res
}
