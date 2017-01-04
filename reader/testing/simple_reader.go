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

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	"golang.org/x/net/context/ctxhttp"
)

// SimpleReader is useful for testing purposes.
type SimpleReader struct {
	name     string
	typeName string
	endpoint string
	mapper   datatype.Mapper
	log      logrus.FieldLogger
	interval time.Duration
	timeout  time.Duration
	backoff  int
	strike   int
	ReadFunc func(context.Context) (*reader.ReadJobResult, error)
	Pinged   bool
}

// NewSimpleReader is a reader for using in tests
func NewSimpleReader(log logrus.FieldLogger, endpoint string, name, typeName string, interval, timeout time.Duration, backoff int) (*SimpleReader, error) {
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

	w := &SimpleReader{
		name:     name,
		typeName: typeName,
		endpoint: url,
		mapper:   &datatype.MapConvertMock{},
		log:      log,
		timeout:  timeout,
		interval: interval,
		backoff:  backoff,
	}
	return w, nil
}

// Ping pings the endpoint and return nil if was successful.
func (s *SimpleReader) Ping() error {
	if s.Pinged {
		// In tests, we have a strict policy on channels. Therefore if it
		// is already pinged, we won't bother.
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	_, err := ctxhttp.Head(ctx, nil, s.endpoint)
	if err != nil {
		return reader.ErrEndpointNotAvailable{Endpoint: s.endpoint, Err: err}
	}
	s.Pinged = true
	return nil
}

// Read executes the ReadFunc if defined, otherwise continues normally
func (s *SimpleReader) Read(job context.Context) (*reader.ReadJobResult, error) {
	if !s.Pinged {
		return nil, reader.ErrPingNotCalled
	}

	if s.strike > s.backoff {
		return nil, reader.ErrBackoffExceeded
	}
	if s.ReadFunc != nil {
		return s.ReadFunc(job)
	}
	id := communication.JobValue(job)
	resp, err := ctxhttp.Get(job, nil, s.endpoint)
	if err != nil {
		if v, ok := err.(*url.Error); ok {
			if strings.Contains(v.Error(), "getsockopt: connection refused") {
				s.strike++
			}
			err = reader.ErrEndpointNotAvailable{Endpoint: s.endpoint, Err: err}
		}
		return nil, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	res := &reader.ReadJobResult{
		ID:       id,
		Time:     time.Now(),
		Res:      buf.Bytes(),
		TypeName: s.TypeName(),
		Mapper:   s.Mapper(),
	}
	return res, nil
}

// Name returns the name
func (s *SimpleReader) Name() string { return s.name }

// TypeName returns the type name
func (s *SimpleReader) TypeName() string { return s.typeName }

// Mapper returns the mapper
func (s *SimpleReader) Mapper() datatype.Mapper { return s.mapper }

// Interval returns the interval
func (s *SimpleReader) Interval() time.Duration { return s.interval }

// Timeout returns the timeout
func (s *SimpleReader) Timeout() time.Duration { return s.timeout }
