// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"bytes"
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/token"
	"github.com/pkg/errors"

	"golang.org/x/net/context/ctxhttp"
)

// Reader is useful for testing purposes.
type Reader struct {
	name     string
	typeName string
	endpoint string
	mapper   datatype.Mapper
	log      internal.FieldLogger
	interval time.Duration
	timeout  time.Duration
	backoff  int
	strike   int
	ReadFunc func(*token.Context) (*reader.Result, error)
	Pinged   bool
}

// New is a reader for using in tests
func New(options ...func(reader.Constructor) error) (*Reader, error) {
	r := &Reader{}
	for _, op := range options {
		err := op(r)
		if err != nil {
			return nil, errors.Wrap(err, "option creation")
		}
	}

	if r.log == nil {
		r.log = internal.GetLogger("info")
	}
	r.log = r.log.WithField("engine", "reader_testing")

	if r.name == "" {
		return nil, reader.ErrEmptyName
	}

	if r.backoff < 5 {
		r.backoff = 5
	}

	if r.mapper == nil {
		r.mapper = &datatype.MapConvertMock{}
	}

	if r.typeName == "" {
		r.typeName = r.name
	}

	if r.interval == 0 {
		r.interval = time.Second
	}

	if r.timeout == 0 {
		r.timeout = 5 * time.Second
	}

	return r, nil
}

// Ping pings the endpoint and return nil if was successful.
func (r *Reader) Ping() error {
	if r.Pinged {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	_, err := ctxhttp.Head(ctx, nil, r.endpoint)
	if err != nil {
		return reader.ErrEndpointNotAvailable{Endpoint: r.endpoint, Err: err}
	}
	r.Pinged = true
	return nil
}

// Read executes the ReadFunc if defined, otherwise continues normally
func (r *Reader) Read(job *token.Context) (*reader.Result, error) {
	if !r.Pinged {
		return nil, reader.ErrPingNotCalled
	}
	if r.strike > r.backoff {
		return nil, reader.ErrBackoffExceeded
	}
	if r.ReadFunc != nil {
		return r.ReadFunc(job)
	}
	resp, err := ctxhttp.Get(job, nil, r.endpoint)
	if err != nil {
		if v, ok := err.(*url.Error); ok {
			if strings.Contains(v.Error(), "getsockopt: connection refused") {
				r.strike++
			}
			err = reader.ErrEndpointNotAvailable{Endpoint: r.endpoint, Err: err}
		}
		return nil, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	res := &reader.Result{
		ID:       job.ID(),
		Time:     time.Now(),
		Content:  buf.Bytes(),
		TypeName: r.TypeName(),
		Mapper:   r.Mapper(),
	}
	return res, nil
}

// Name returns the name
func (r *Reader) Name() string { return r.name }

// SetName sets the name of the reader
func (r *Reader) SetName(name string) { r.name = name }

// Endpoint returns the endpoint
func (r *Reader) Endpoint() string { return r.endpoint }

// SetEndpoint sets the endpoint of the reader
func (r *Reader) SetEndpoint(endpoint string) { r.endpoint = endpoint }

// TypeName returns the type name
func (r *Reader) TypeName() string { return r.typeName }

// SetTypeName sets the type name of the reader
func (r *Reader) SetTypeName(typeName string) { r.typeName = typeName }

// Mapper returns the mapper
func (r *Reader) Mapper() datatype.Mapper { return r.mapper }

// SetMapper sets the mapper of the reader
func (r *Reader) SetMapper(mapper datatype.Mapper) { r.mapper = mapper }

// Interval returns the interval
func (r *Reader) Interval() time.Duration { return r.interval }

// SetInterval sets the interval of the reader
func (r *Reader) SetInterval(interval time.Duration) { r.interval = interval }

// Timeout returns the timeout
func (r *Reader) Timeout() time.Duration { return r.timeout }

// SetTimeout sets the timeout of the reader
func (r *Reader) SetTimeout(timeout time.Duration) { r.timeout = timeout }

// Backoff returns the backoff
func (r *Reader) Backoff() int { return r.backoff }

// SetBackoff sets the backoff of the reader
func (r *Reader) SetBackoff(backoff int) { r.backoff = backoff }

// Logger returns the log
func (r *Reader) Logger() internal.FieldLogger { return r.log }

// SetLogger sets the log of the reader
func (r *Reader) SetLogger(log internal.FieldLogger) { r.log = log }
