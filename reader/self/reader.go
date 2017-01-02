// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package self contains codes for recording expvastic's own metrics.
//
// Collected metrics
//
// This list will grow in time:
//      ElasticSearch Var Name    | expvastic var name
//      ----------------------------------------------
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
	"bytes"
	"context"
	"expvar"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
)

// Reader reads from expvastic own application's metric information.
// It implements DataReader interface.
type Reader struct {
	name     string
	typeName string
	log      logrus.FieldLogger
	mapper   datatype.Mapper
	interval time.Duration
	timeout  time.Duration
	backoff  int
	strike   int
	quit     chan struct{}
	endpoint string
}

// New exposes expvastic's own metrics.
// It returns and error on the following occasions:
//
//   Condition      |  Error
//   ---------------|-------------
//   name == ""     | ErrEmptyName
//   typeName == "" | ErrEmptyTypeName
//   backoff < 5    | ErrLowBackoffValue
//
func New(log logrus.FieldLogger, endpoint string, mapper datatype.Mapper, name, typeName string, interval time.Duration, timeout time.Duration, backoff int) (*Reader, error) {
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

	if backoff < 5 {
		return nil, reader.ErrLowBackoffValue(backoff)
	}

	log = log.WithField("engine", "expvastic")
	w := &Reader{
		name:     name,
		typeName: typeName,
		mapper:   mapper,
		log:      log,
		interval: interval,
		timeout:  timeout,
		endpoint: url,
		backoff:  backoff,
		quit:     make(chan struct{}),
	}
	return w, nil
}

// Read send the metrics back. The error is usually nil.
func (r *Reader) Read(job context.Context) (*reader.ReadJobResult, error) {
	if r.endpoint != IgnoredEndpoint {
		// To support the tests
		return r.readMetricsFromURL(job)
	}

	id := communication.JobValue(job)
	buf := new(bytes.Buffer) // construct a json encoder and pass it
	fmt.Fprintf(buf, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(buf, ",\n")
		}
		first = false
		fmt.Fprintf(buf, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(buf, "\n}\n")
	res := &reader.ReadJobResult{
		ID:       id,
		Time:     time.Now(), // It is sensible to record the time now
		Res:      buf.Bytes(),
		TypeName: r.TypeName(),
		Mapper:   r.Mapper(),
	}
	return res, nil
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

func (r *Reader) readMetricsFromURL(job context.Context) (*reader.ReadJobResult, error) {
	if r.strike > r.backoff {
		return nil, reader.ErrBackoffExceeded
	}

	id := communication.JobValue(job)
	resp, err := http.Get(r.endpoint)
	if err != nil {
		if v, ok := err.(*url.Error); ok {
			if strings.Contains(v.Error(), "getsockopt: connection refused") {
				r.strike++
			}
		}
		r.log.WithField("reader", "self").
			WithField("ID", id).
			Errorf("%s: error making request: %v", r.name, err) // Error because it is a local dependency.
		return nil, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	res := &reader.ReadJobResult{
		ID:       id,
		Time:     time.Now(), // It is sensible to record the time now
		Res:      buf.Bytes(),
		TypeName: r.TypeName(),
		Mapper:   r.Mapper(),
	}
	return res, nil
}
