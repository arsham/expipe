// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package elasticsearch contains logic to record data to an elasticsearch index.
// The data is already sanitised by the data provider.
package elasticsearch

import (
	"context"
	"expvar"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/recorder"
	"github.com/olivere/elastic"
	"github.com/pkg/errors"
)

var elasticsearchRecords = expvar.NewInt("ElasticSearch Records")

// Recorder contains an elasticsearch client and an index name for recording data
// It implements DataRecorder interface
type Recorder struct {
	name      string
	client    *elastic.Client // Elasticsearch client
	endpoint  string
	indexName string
	log       logrus.FieldLogger
	timeout   time.Duration
	backoff   int
	strike    int
	pinged    bool
}

// New returns an error if it can't create the index
// It returns and error on the following occasions:
//
//   Condition            |  Error
//   ---------------------|-------------
//   Invalid endpoint     | ErrInvalidEndpoint
//   backoff < 5          | ErrLowBackoffValue
//   Empty name           | ErrEmptyName
//   Invalid IndexName    | ErrInvalidIndexName
//   Empty IndexName      | ErrEmptyIndexName
//
func New(ctx context.Context, log logrus.FieldLogger, name, endpoint, indexName string, timeout time.Duration, backoff int) (*Recorder, error) {
	if name == "" {
		return nil, recorder.ErrEmptyName
	}

	if indexName == "" {
		return nil, recorder.ErrEmptyIndexName
	}

	if strings.ContainsAny(indexName, ` "*\<|,>/?`) {
		return nil, recorder.ErrInvalidIndexName(indexName)
	}

	log.Debug("connecting to: ", endpoint)
	url, err := lib.SanitiseURL(endpoint)
	if err != nil {
		return nil, recorder.ErrInvalidEndpoint(endpoint)
	}
	if backoff < 5 {
		return nil, recorder.ErrLowBackoffValue(backoff)
	}

	return &Recorder{
		name:      name,
		endpoint:  url,
		indexName: indexName,
		log:       log,
		timeout:   timeout,
		backoff:   backoff,
	}, nil
}

// Ping should ping the endpoint and report if was successful.
// It returns and error on the following occasions:
//
//   Condition            |  Error
//   ---------------------|-------------
//   Unavailable endpoint | ErrEndpointNotAvailable
//   Ping errors          | Timeout/Ping failed
//   Index creation       | elasticsearch's errors
//
func (r *Recorder) Ping() error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	addr := elastic.SetURL(r.endpoint)
	logger := elastic.SetErrorLog(r.log)

	r.client, err = elastic.NewClient(
		addr,
		logger,
		elastic.SetHealthcheckTimeoutStartup(r.timeout),
		elastic.SetSnifferTimeout(r.timeout),
		elastic.SetHealthcheckTimeout(r.timeout),
	)
	if err != nil {
		return recorder.ErrEndpointNotAvailable{Endpoint: r.endpoint, Err: err}
	}
	_, _, err = r.client.Ping(r.endpoint).Do(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return errors.Wrapf(err, "timeout: %s", ctx.Err())
		}
		return errors.Wrap(err, "ping failed")
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
func (r *Recorder) Record(ctx context.Context, job *recorder.Job) error {
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
		if err == elastic.ErrNoClient {
			r.strike++
			err = recorder.ErrEndpointNotAvailable{Endpoint: r.endpoint, Err: err}
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
// otherwise continues as normal.
// Although this doesn't change the state of the Client, it is a part of its behaviour.
func (r *Recorder) record(ctx context.Context, typeName string, timestamp time.Time, list datatype.DataContainer) error {
	payload := string(list.Bytes(timestamp))
	_, err := r.client.Index().
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

// Name shows the name identifier for this reader
func (r *Recorder) Name() string { return r.name }

// IndexName is the index/database
func (r *Recorder) IndexName() string { return r.indexName }

// Timeout returns the timeout
func (r *Recorder) Timeout() time.Duration { return r.timeout }
