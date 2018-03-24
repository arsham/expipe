// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/recorder"
	"github.com/pkg/errors"
	"github.com/shurcooL/go/ctxhttp"
)

// Recorder is designed to be used in tests.
type Recorder struct {
	name       string
	endpoint   string
	indexName  string
	log        internal.FieldLogger
	timeout    time.Duration
	ErrorFunc  func() error
	backoff    int
	strike     int
	Smu        sync.RWMutex
	RecordFunc func(context.Context, *recorder.Job) error
	Pinged     bool
}

// New is a recorder for using in tests
func New(options ...func(recorder.Constructor) error) (*Recorder, error) {
	r := &Recorder{}
	for _, op := range options {
		err := op(r)
		if err != nil {
			return nil, errors.Wrap(err, "option creation")
		}
	}
	if r.log == nil {
		r.log = internal.GetLogger("error")
	}
	r.log = r.log.WithField("engine", "recorder_testing")
	if r.backoff < 5 {
		r.backoff = 5
	}
	if r.indexName == "" {
		r.indexName = r.name
	}
	if r.timeout == 0 {
		r.timeout = 5 * time.Second
	}
	return r, nil
}

// Ping pings the endpoint and return nil if was successful.
func (r *Recorder) Ping() error {
	if r.Pinged {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	_, err := ctxhttp.Head(ctx, nil, r.endpoint)
	if err != nil {
		return recorder.ErrEndpointNotAvailable{Endpoint: r.endpoint, Err: err}
	}
	r.Pinged = true
	return nil
}

// Record calls the RecordFunc if exists, otherwise continues as normal
func (r *Recorder) Record(ctx context.Context, job *recorder.Job) error {
	if !r.Pinged {
		return recorder.ErrPingNotCalled
	}
	r.Smu.RLock()
	if r.RecordFunc != nil {
		r.Smu.RUnlock()
		return r.RecordFunc(ctx, job)
	}
	r.Smu.RUnlock()

	if r.strike > r.backoff {
		return recorder.ErrBackoffExceeded
	}

	res, err := http.Get(r.endpoint)
	if err != nil {
		if v, ok := err.(*url.Error); ok {
			if strings.Contains(v.Error(), "getsockopt: connection refused") {
				r.strike++
			}
		}
		return err
	}
	res.Body.Close()
	return nil
}

// Name returns the name
func (r *Recorder) Name() string { return r.name }

// SetName sets the name of the recorder
func (r *Recorder) SetName(name string) { r.name = name }

// Endpoint returns the endpoint
func (r *Recorder) Endpoint() string { return r.endpoint }

// SetEndpoint sets the endpoint of the recorder
func (r *Recorder) SetEndpoint(endpoint string) { r.endpoint = endpoint }

// IndexName returns the index name
func (r *Recorder) IndexName() string { return r.indexName }

// SetIndexName sets the index name of the recorder
func (r *Recorder) SetIndexName(indexName string) { r.indexName = indexName }

// Timeout returns the timeout
func (r *Recorder) Timeout() time.Duration { return r.timeout }

// SetTimeout sets the timeout of the recorder
func (r *Recorder) SetTimeout(timeout time.Duration) { r.timeout = timeout }

// Backoff returns the backoff
func (r *Recorder) Backoff() int { return r.backoff }

// SetBackoff sets the backoff of the recorder
func (r *Recorder) SetBackoff(backoff int) { r.backoff = backoff }

// Logger returns the log
func (r *Recorder) Logger() internal.FieldLogger { return r.log }

// SetLogger sets the log of the recorder
func (r *Recorder) SetLogger(log internal.FieldLogger) { r.log = log }
