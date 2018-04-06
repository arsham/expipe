// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"bytes"
	"context"
	"net/url"
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/tools"
	"github.com/arsham/expipe/tools/token"
	"github.com/pkg/errors"

	"golang.org/x/net/context/ctxhttp"
)

// Reader is useful for testing purposes.
type Reader struct {
	MockName     string
	MockTypeName string
	MockEndpoint string
	MockMapper   datatype.Mapper
	log          tools.FieldLogger
	MockInterval time.Duration
	timeout      time.Duration
	ReadFunc     func(*token.Context) (*reader.Result, error)
	PingFunc     func() error
	Pinged       bool
}

// New is a reader for using in tests.
func New(options ...func(reader.Constructor) error) (*Reader, error) {
	r := &Reader{}
	for _, op := range options {
		err := op(r)
		if err != nil {
			return nil, errors.Wrap(err, "option creation")
		}
	}
	if err := checkReader(r); err != nil {
		return nil, err
	}
	r.log = r.log.WithField("engine", "reader_testing")
	return r, nil
}

func checkReader(r *Reader) error {
	if r.MockName == "" {
		return reader.ErrEmptyName
	}
	if r.MockEndpoint == "" {
		return reader.ErrEmptyEndpoint
	}
	if r.MockMapper == nil {
		r.MockMapper = &datatype.MapConvertMock{}
	}
	if r.MockTypeName == "" {
		r.MockTypeName = r.MockName
	}
	if r.MockInterval == 0 {
		r.MockInterval = time.Second
	}
	if r.timeout == 0 {
		r.timeout = 5 * time.Second
	}
	if r.log == nil {
		r.log = tools.GetLogger("info")
	}
	return nil
}

// Ping pings the endpoint and return nil if was successful.
func (r *Reader) Ping() error {
	if r.PingFunc != nil {
		return r.PingFunc()
	}
	if r.Pinged {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	_, err := ctxhttp.Head(ctx, nil, r.MockEndpoint)
	if err != nil {
		return reader.EndpointNotAvailableError{
			Endpoint: r.MockEndpoint,
			Err:      err,
		}
	}
	r.Pinged = true
	return nil
}

// Read executes the ReadFunc if defined, otherwise continues normally.
func (r *Reader) Read(job *token.Context) (*reader.Result, error) {
	if r.ReadFunc != nil {
		return r.ReadFunc(job)
	}
	if !r.Pinged {
		return nil, reader.ErrPingNotCalled
	}
	resp, err := ctxhttp.Get(job, nil, r.MockEndpoint)
	if err != nil {
		if _, ok := err.(*url.Error); ok {
			err = reader.EndpointNotAvailableError{
				Endpoint: r.MockEndpoint,
				Err:      err,
			}
		}
		return nil, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if !tools.IsJSON(buf.Bytes()) {
		return nil, reader.ErrInvalidJSON
	}
	res := &reader.Result{
		ID:       job.ID(),
		Time:     time.Now(),
		Content:  buf.Bytes(),
		TypeName: r.TypeName(),
		Mapper:   r.Mapper(),
	}
	return res, nil
}

// Name returns the name.
func (r *Reader) Name() string { return r.MockName }

// SetName sets the name of the reader.
func (r *Reader) SetName(name string) { r.MockName = name }

// Endpoint returns the endpoint.
func (r *Reader) Endpoint() string { return r.MockEndpoint }

// SetEndpoint sets the endpoint of the reader.
func (r *Reader) SetEndpoint(endpoint string) { r.MockEndpoint = endpoint }

// TypeName returns the type name.
func (r *Reader) TypeName() string { return r.MockTypeName }

// SetTypeName sets the type name of the reader.
func (r *Reader) SetTypeName(typeName string) { r.MockTypeName = typeName }

// Mapper returns the mapper.
func (r *Reader) Mapper() datatype.Mapper { return r.MockMapper }

// SetMapper sets the mapper of the reader.
func (r *Reader) SetMapper(mapper datatype.Mapper) { r.MockMapper = mapper }

// Interval returns the interval.
func (r *Reader) Interval() time.Duration { return r.MockInterval }

// SetInterval sets the interval of the reader.
func (r *Reader) SetInterval(interval time.Duration) { r.MockInterval = interval }

// Timeout returns the timeout.
func (r *Reader) Timeout() time.Duration { return r.timeout }

// SetTimeout sets the timeout of the reader.
func (r *Reader) SetTimeout(timeout time.Duration) { r.timeout = timeout }

// Logger returns the log.
func (r *Reader) Logger() tools.FieldLogger { return r.log }

// SetLogger sets the log of the reader.
func (r *Reader) SetLogger(log tools.FieldLogger) { r.log = log }
