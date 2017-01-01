// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"bytes"
	"context"
	"net/http"
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
}

// NewSimpleReader is a reader for using in tests
func NewSimpleReader(
	log logrus.FieldLogger,
	endpoint string,
	name,
	typeName string,
	interval,
	timeout time.Duration,
	backoff int,
) (*SimpleReader, error) {
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

// Read executes the ReadFunc if defined, otherwise continues normally
func (m *SimpleReader) Read(job context.Context) (*reader.ReadJobResult, error) {
	if m.strike > m.backoff {
		return nil, reader.ErrBackoffExceeded
	}
	if m.ReadFunc != nil {
		return m.ReadFunc(job)
	}
	id := communication.JobValue(job)
	resp, err := ctxhttp.Get(job, nil, m.endpoint)
	if err != nil {
		if v, ok := err.(*url.Error); ok {
			if strings.Contains(v.Error(), "getsockopt: connection refused") {
				m.strike++
			}
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
		TypeName: m.TypeName(),
		Mapper:   m.Mapper(),
	}
	return res, nil
}

// Name returns the name
func (m *SimpleReader) Name() string { return m.name }

// TypeName returns the type name
func (m *SimpleReader) TypeName() string { return m.typeName }

// Mapper returns the mapper
func (m *SimpleReader) Mapper() datatype.Mapper { return m.mapper }

// Interval returns the interval
func (m *SimpleReader) Interval() time.Duration { return m.interval }

// Timeout returns the timeout
func (m *SimpleReader) Timeout() time.Duration { return m.timeout }
