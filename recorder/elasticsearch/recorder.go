// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package elasticsearch

import (
    "context"
    "fmt"
    "time"

    "github.com/Sirupsen/logrus"
    "github.com/arsham/expvastic/datatype"
    "github.com/arsham/expvastic/recorder"
    "github.com/olivere/elastic"
)

// ElasticSearch contains an elasticsearch client and an indexname for recording data
// It implements DataRecorder interface
type ElasticSearch struct {
    client    *elastic.Client // ElasticSearch client
    indexName string
    jobChan   chan *recorder.RecordJob
}

// NewElasticSearch returns an error if it can't create the index
func NewElasticSearch(bgCtx context.Context, log logrus.FieldLogger, esURL, indexName string) (*ElasticSearch, error) {
    addr := elastic.SetURL(esURL)
    logger := elastic.SetErrorLog(log)
    client, err := elastic.NewClient(addr, logger)
    if err != nil {
        log.Fatal(err)
    }

    // QUESTION: Is there any significant for this cancel?
    ctx, _ := context.WithTimeout(bgCtx, 10*time.Second)
    _, _, err = client.Ping(esURL).Do(ctx)
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
    return &ElasticSearch{
        client:    client,
        indexName: indexName,
        jobChan:   make(chan *recorder.RecordJob),
    }, nil
}

// Start begins reading from the target in its own goroutine
// It will close the done channel when the job channel is closed
func (e *ElasticSearch) Start() chan struct{} {
    done := make(chan struct{})
    go func() {
        for job := range e.jobChan {
            go func(job *recorder.RecordJob) {
                job.Err <- e.record(job.Ctx, job.TypeName, job.Time, job.Payload)
            }(job)
        }
        close(done)
    }()
    return done
}

// PayloadChan returns the channel it receives the information from
func (e *ElasticSearch) PayloadChan() chan *recorder.RecordJob {
    return e.jobChan
}

// record ships the kv data to elasticsearch
// Although this doesn't change the state of the Client, it is a part of its behaviour
func (e *ElasticSearch) record(ctx context.Context, typeName string, timestamp time.Time, list datatype.DataContainer) error {
    payload := list.String(timestamp)
    _, err := e.client.Index().
        Index(e.indexName).
        Type(typeName).
        BodyString(payload).
        Do(ctx)
    if err != nil {
        return err
    }
    return ctx.Err()
}
