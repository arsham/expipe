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
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/recorder"
	"github.com/olivere/elastic"
)

var elasticsearchRecords = expvar.NewInt("ElasticSearch Records")

// Recorder contains an elasticsearch client and an index name for recording data
// It implements DataRecorder interface
type Recorder struct {
	name        string
	client      *elastic.Client // Elasticsearch client
	indexName   string
	payloadChan chan *recorder.RecordJob
	errorChan   chan<- communication.ErrorMessage
	log         logrus.FieldLogger
	timeout     time.Duration
}

// NewRecorder returns an error if it can't create the index
func NewRecorder(
	ctx context.Context,
	log logrus.FieldLogger,
	payloadChan chan *recorder.RecordJob,
	errorChan chan<- communication.ErrorMessage,
	name,
	endpoint,
	indexName string,
	timeout time.Duration,
) (*Recorder, error) {
	log.Debug("connecting to: ", endpoint)
	addr := elastic.SetURL(endpoint)
	logger := elastic.SetErrorLog(log)
	client, err := elastic.NewClient(addr, logger)
	if err != nil {
		log.Fatal(err)
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
	return &Recorder{
		name:        name,
		client:      client,
		indexName:   indexName,
		payloadChan: payloadChan,
		errorChan:   errorChan,
		log:         log,
		timeout:     timeout,
	}, nil
}

// Start begins reading from the target in its own goroutine
// It will close the done channel when the job channel is closed
func (r *Recorder) Start(ctx context.Context, stop communication.StopChannel) {
	go func() {
		for {
			select {
			case job := <-r.payloadChan:
				go func(job *recorder.RecordJob) {
					err := r.record(job.Ctx, job.TypeName, job.Time, job.Payload)
					if err != nil {
						r.errorChan <- communication.ErrorMessage{ID: job.ID, Name: r.Name(), Err: err}
					}
				}(job)
			case s := <-stop:
				// TODO: make a condition here, we don't want to lose the data that is happening before we quit.
				s <- struct{}{}
				return
			}
		}
	}()
}

// PayloadChan returns the channel it receives the information from
func (r *Recorder) PayloadChan() chan *recorder.RecordJob { return r.payloadChan }

// record ships the kv data to elasticsearch
// Although this doesn't change the state of the Client, it is a part of its behaviour
func (r *Recorder) record(ctx context.Context, typeName string, timestamp time.Time, list datatype.DataContainer) error {
	payload := list.String(timestamp)
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
