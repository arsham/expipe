// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
    "context"
    "fmt"
    "time"

    "github.com/Sirupsen/logrus"
    "github.com/olivere/elastic"
)

// ElasticSearch ...
type ElasticSearch struct {
    client    *elastic.Client // ElasticSearch client
    indexName string
}

// NewElasticSearch returns an error if it can't create the index
func NewElasticSearch(bgCtx context.Context, log logrus.FieldLogger, esURL, indexName string) (*ElasticSearch, error) {
    addr := elastic.SetURL(esURL)
    logger := elastic.SetErrorLog(log)
    client, err := elastic.NewClient(addr, logger)
    if err != nil {
        log.Fatal(err)
    }

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
    }, nil
}

// Record ships the kv data to elasticsearch
// Although this doesn't change the state of the Client, it is a part of its behaviour
func (e *ElasticSearch) Record(ctx context.Context, typeName string, timestamp time.Time, kv []DataType) error {
    payload := getQueryString(timestamp, kv)
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
