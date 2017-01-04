// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package expvar contains logic to read from an expvar provide. The data comes in JSON format. The GC
// and memory related information will be changed to better presented to the data recorders.
// Bytes will be turned into megabytes, gc lists will be truncated to remove zero values.
package expvar

import (
	"bytes"
	"context"
	"expvar"
	"net/url"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	"golang.org/x/net/context/ctxhttp"
)

var (
	expvarReads = expvar.NewInt("Expvar Reads")
)

// Reader can read from any application that exposes expvar information.
// It implements DataReader interface.
type Reader struct {
	name     string
	endpoint string
	log      logrus.FieldLogger
	mapper   datatype.Mapper
	typeName string
	interval time.Duration
	timeout  time.Duration
	backoff  int
	strike   int
	pinged   bool
}

// New creates the worker and sets up its channels.
// It returns and error on the following occasions:
//
//   Condition            |  Error
//   ---------------------|-------------
//   name == ""           | ErrEmptyName
//   endpoint == ""       | ErrEmptyEndpoint
//   Invalid endpoint     | ErrInvalidEndpoint
//   Unavailable endpoint | ErrEndpointNotAvailable
//   typeName == ""       | ErrEmptyTypeName
//   backoff < 5          | ErrLowBackoffValue
//
func New(log logrus.FieldLogger, endpoint string, mapper datatype.Mapper, name string, typeName string, interval time.Duration, timeout time.Duration, backoff int) (*Reader, error) {
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

	if typeName == "" {
		return nil, reader.ErrEmptyTypeName
	}

	if backoff < 5 {
		return nil, reader.ErrLowBackoffValue(backoff)
	}

	log = log.WithField("reader", "expvar").WithField("name", name)
	w := &Reader{
		name:     name,
		typeName: typeName,
		mapper:   mapper,
		endpoint: url,
		log:      log,
		timeout:  timeout,
		interval: interval,
		backoff:  backoff,
	}
	return w, nil
}

// Ping pings the endpoint and return nil if was successful.
func (r *Reader) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	_, err := ctxhttp.Head(ctx, nil, r.endpoint)
	if err != nil {
		return reader.ErrEndpointNotAvailable{Endpoint: r.endpoint, Err: err}
	}
	r.pinged = true
	return nil
}

// Read begins reading from the target.
// It returns an error back to the engine if it can't read from metrics provider.
func (r *Reader) Read(job context.Context) (*reader.ReadJobResult, error) {
	if !r.pinged {
		return nil, reader.ErrPingNotCalled
	}
	if r.strike > r.backoff {
		return nil, reader.ErrBackoffExceeded
	}
	id := communication.JobValue(job)
	resp, err := ctxhttp.Get(job, nil, r.endpoint)

	if err != nil {
		if v, ok := err.(*url.Error); ok {
			if strings.Contains(v.Error(), "getsockopt: connection refused") {
				r.strike++
			}
			err = reader.ErrEndpointNotAvailable{Endpoint: r.endpoint, Err: err}
		}
		r.log.WithField("reader", "expvar_reader").
			WithField("name", r.Name()).
			WithField("ID", id).
			Debugf("%s: error making request: %v", r.name, err)
		return nil, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	res := &reader.ReadJobResult{
		ID:       id,
		Time:     time.Now(), // It is sensible to record the time now
		Res:      buf.Bytes(),
		TypeName: r.TypeName(),
		Mapper:   r.Mapper(),
	}
	expvarReads.Add(1)
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

// Timeout returns the time-out
func (r *Reader) Timeout() time.Duration { return r.timeout }
