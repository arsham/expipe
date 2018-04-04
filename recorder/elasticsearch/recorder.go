// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package elasticsearch contains logic to record data to an elasticsearch index.
// The data is already sanitised by the data provider.
package elasticsearch

import (
	"bytes"
	"context"
	"expvar"
	"net/url"
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/recorder"
	"github.com/arsham/expipe/tools"
	"github.com/olivere/elastic"
	"github.com/pkg/errors"
)

var elasticsearchRecords = expvar.NewInt("ElasticSearch Records")

// Recorder contains an elasticsearch client and an index name for recording
// data. It implements DataRecorder interface
type Recorder struct {
	name      string
	client    *elastic.Client // Elasticsearch client
	endpoint  string
	indexName string
	log       tools.FieldLogger
	timeout   time.Duration
	backoff   int
	strike    int
	pinged    bool
}

// New returns an error if it can't create the index
// It returns and error on the following occasions:
//
//   +-------------------+-----------------------+
//   |     Condition     |         Error         |
//   +-------------------+-----------------------+
//   | Invalid endpoint  | InvalidEndpointError  |
//   | backoff < 5       | LowBackoffValueError  |
//   | Empty name        | ErrEmptyName          |
//   | Invalid IndexName | InvalidIndexNameError |
//   | Empty IndexName   | ErrEmptyIndexName     |
//   +-------------------+-----------------------+
//
func New(options ...func(recorder.Constructor) error) (*Recorder, error) {
	r := &Recorder{}
	for _, op := range options {
		err := op(r)
		if err != nil {
			return nil, errors.Wrap(err, "option creation")
		}
	}
	if r.name == "" {
		return nil, recorder.ErrEmptyName
	}
	if r.endpoint == "" {
		return nil, recorder.ErrEmptyEndpoint
	}
	if r.log == nil {
		r.log = tools.GetLogger("error")
	}
	r.log = r.log.WithField("engine", "expipe")
	if r.backoff < 5 {
		r.backoff = 5
	}
	if r.indexName == "" {
		r.indexName = r.name
	}
	if r.timeout == 0 {
		r.timeout = 5 * time.Second
	}
	r.log.Debug("connecting to: ", r.Endpoint())
	return r, nil
}

// Ping should ping the endpoint and report if was successful.
// It returns and error on the following occasions:
//
//   +----------------------+---------------------------+
//   |      Condition       |           Error           |
//   +----------------------+---------------------------+
//   | Unavailable endpoint | EndpointNotAvailableError |
//   | Ping errors          | Timeout/Ping failed       |
//   | Index creation       | elasticsearch's errors    |
//   +----------------------+---------------------------+
//
func (r *Recorder) Ping() error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	r.client, err = elastic.NewClient(
		elastic.SetURL(r.endpoint),
		elastic.SetErrorLog(r.log),
		elastic.SetHealthcheckTimeoutStartup(r.timeout),
		elastic.SetSnifferTimeout(r.timeout),
		elastic.SetHealthcheckTimeout(r.timeout),
	)
	if err != nil {
		return recorder.EndpointNotAvailableError{Endpoint: r.endpoint, Err: err}
	}
	exists, err := r.client.IndexExists(r.indexName).Do(ctx)
	if err != nil {
		return errors.Wrap(err, "querying index")
	}
	if !exists {
		_, err := r.client.CreateIndex(r.indexName).Do(ctx)
		if err != nil {
			return errors.Wrapf(err, "create index: %s", r.indexName)
		}
	}
	r.pinged = true
	return nil
}

// Record returns an error if the endpoint responds in errors. It stops
// receiving jobs when the endpoint's absence has exceeded the backoff value.
// It returns an error if the ping is not called or the endpoint
// is not responding too many times.
func (r *Recorder) Record(ctx context.Context, job recorder.Job) error {
	if !r.pinged {
		return recorder.ErrPingNotCalled
	}
	if r.strike > r.backoff {
		return recorder.ErrBackoffExceeded
	}
	ctx, cancel := context.WithTimeout(ctx, r.Timeout())
	defer cancel()
	err := r.record(ctx, job.TypeName, job.Time, job.Payload)
	if err != nil {
		err = errors.Cause(err)
		if _, ok := err.(*url.Error); ok || err == elastic.ErrNoClient {
			r.strike++
			err = recorder.EndpointNotAvailableError{Endpoint: r.endpoint, Err: err}
		}
		r.log.WithField("recorder", "elasticsearch").
			WithField("name", r.Name()).
			WithField("ID", job.ID).
			Debugf("%s: error making request: %v", r.name, err)
		return err
	}

	return nil
}

// record ships the kv data to elasticsearch. It calls the recordFunc if exists,
// otherwise continues as normal. Although this doesn't change the state of
// the Client, it is a part of its behaviour.
func (r *Recorder) record(ctx context.Context, typeName string, timestamp time.Time, list datatype.DataContainer) error {
	w := new(bytes.Buffer)
	_, err := list.Generate(w, timestamp)
	if err != nil {
		errors.Wrap(err, "generating payload")
	}
	payload := w.String()
	_, err = r.client.Index().
		Index(r.indexName).
		Type(typeName).
		BodyString(payload).
		Do(ctx)
	if err != nil {
		return errors.Wrap(err, "record payload")
	}
	elasticsearchRecords.Add(1)
	return ctx.Err()
}

// Name shows the name identifier for this recorder
func (r *Recorder) Name() string { return r.name }

// SetName sets the name of the recorder
func (r *Recorder) SetName(name string) { r.name = name }

// Endpoint returns the endpoint
func (r *Recorder) Endpoint() string { return r.endpoint }

// SetEndpoint sets the endpoint of the recorder
func (r *Recorder) SetEndpoint(endpoint string) { r.endpoint = endpoint }

// IndexName shows the indexName the recorder should record as
func (r *Recorder) IndexName() string { return r.indexName }

// SetIndexName sets the type name of the recorder
func (r *Recorder) SetIndexName(indexName string) { r.indexName = indexName }

// Timeout returns the time-out
func (r *Recorder) Timeout() time.Duration { return r.timeout }

// SetTimeout sets the timeout of the recorder
func (r *Recorder) SetTimeout(timeout time.Duration) { r.timeout = timeout }

// Backoff returns the backoff
func (r *Recorder) Backoff() int { return r.backoff }

// SetBackoff sets the backoff of the recorder
func (r *Recorder) SetBackoff(backoff int) { r.backoff = backoff }

// SetLogger sets the log of the recorder
func (r *Recorder) SetLogger(log tools.FieldLogger) { r.log = log }
