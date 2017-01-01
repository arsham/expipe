// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package elasticsearch contains logic to record data to an elasticsearch index. The data is already sanitised
// by the data provider.
package elasticsearch

import (
	"context"
	"expvar"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/recorder"
	"github.com/olivere/elastic"
)

var elasticsearchRecords = expvar.NewInt("ElasticSearch Records")

// Recorder contains an elasticsearch client and an index name for recording data
// It implements DataRecorder interface
type Recorder struct {
	name       string
	client     *elastic.Client // Elasticsearch client
	endpoint   string
	indexName  string
	log        logrus.FieldLogger
	timeout    time.Duration
	backoff    int
	strike     int
	recordFunc func(ctx context.Context, typeName string, timestamp time.Time, list datatype.DataContainer) error
}

// NewRecorder returns an error if it can't create the index
func NewRecorder(
	ctx context.Context,
	log logrus.FieldLogger,
	name,
	endpoint,
	indexName string,
	timeout time.Duration,
	backoff int,
) (*Recorder, error) {
	log.Debug("connecting to: ", endpoint)
	url, err := lib.SanitiseURL(endpoint)
	if err != nil {
		return nil, recorder.ErrInvalidEndpoint(endpoint)
	}
	endpoint = url
	addr := elastic.SetURL(endpoint)
	logger := elastic.SetErrorLog(log)

	client, err := elastic.NewClient(
		addr,
		logger,
		elastic.SetHealthcheckTimeoutStartup(timeout),
		elastic.SetSnifferTimeout(timeout),
		elastic.SetHealthcheckTimeout(timeout),
	)
	if err != nil {
		return nil, recorder.ErrEndpointNotAvailable{Endpoint: endpoint, Err: err}
	}

	// QUESTION: Is there any significant for this cancel?
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, _, err = client.Ping(endpoint).Do(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("Timeout: %s - %s", ctx.Err(), err)
		}
		return nil, fmt.Errorf("Ping failed: %s", err)
	}

	exists, err := client.IndexExists(indexName).Do(ctx)
	if err != nil {
		return nil, err
	}

	if !exists {
		_, err := client.CreateIndex(indexName).Do(ctx)
		if err != nil {
			return nil, err
		}
	}

	if backoff < 5 {
		return nil, recorder.ErrLowBackoffValue(backoff)
	}

	return &Recorder{
		name:      name,
		client:    client,
		endpoint:  endpoint,
		indexName: indexName,
		log:       log,
		timeout:   timeout,
		backoff:   backoff,
	}, nil
}

// Record returns an error if the endpoint errors. It stops receiving jobs when the
// endpoint's absence has exceeded the backoff value.
func (r *Recorder) Record(ctx context.Context, job *recorder.RecordJob) error {
	if r.strike > r.backoff {
		return recorder.ErrBackoffExceeded
	}
	err := r.record(ctx, job.TypeName, job.Time, job.Payload)
	if err != nil {
		if v, ok := err.(*url.Error); ok {
			if strings.Contains(v.Error(), "getsockopt: connection refused") {
				r.strike++
			}
		}
		r.log.WithField("recorder", "elasticsearch").
			WithField("name", r.Name()).
			WithField("ID", job.ID).
			Debugf("%s: error making request: %v", r.name, err)
		return err
	}
	return nil
}

// record ships the kv data to elasticsearch. It calls the recordFunc if exists, otherwise continues as normal.
// Although this doesn't change the state of the Client, it is a part of its behaviour
func (r *Recorder) record(ctx context.Context, typeName string, timestamp time.Time, list datatype.DataContainer) error {
	if r.recordFunc != nil {
		return r.recordFunc(ctx, typeName, timestamp, list)
	}
	payload := string(list.Bytes(timestamp))
	_, err := r.client.Index().
		Index(r.indexName).
		Type(typeName).
		BodyString(payload).
		Do(ctx)
	if err != nil {
		return err
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

// SetRecordFunc sets the recordFunc. You should only use it in tests.
func (r *Recorder) SetRecordFunc(f func(ctx context.Context, typeName string, timestamp time.Time, list datatype.DataContainer) error) {
	r.recordFunc = f
}
