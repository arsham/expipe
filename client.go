// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "strings"
    "time"

    "github.com/Sirupsen/logrus"
    "github.com/olivere/elastic"
)

// Client represents a client that can put information into an ES index
// The client is allowed to change the index and type names at will
type Client struct {
    ctx       context.Context // When this context is canceled, it client tries to finalize its work
    client    *elastic.Client // ElasticSearch client
    target    string          // Target endpoint to listen to
    indexName string          // ElasticSearch index name
    typeName  string          //ElasticSearch type name
    interval  time.Duration
    timeout   time.Duration
    logger    logrus.FieldLogger
}

// NewClient creates an index if not exists
// It returns an error if index creation is unsuccessful
func NewClient(ctx context.Context, client *elastic.Client, c Conf) (*Client, error) {
    exists, err := client.IndexExists(c.IndexName).Do(ctx)
    if err != nil {
        return nil, err
    }

    if !exists {
        _, err := client.CreateIndex(c.IndexName).Do(ctx)
        if err != nil {
            return nil, err
        }
    }

    // TODO: ping the target
    return &Client{
        client:    client,
        target:    c.Target,
        indexName: c.IndexName,
        typeName:  c.TypeName,
        interval:  c.Interval,
        timeout:   c.Timeout,
        logger:    c.Logger,
    }, nil
}

// Record ships the kv data to elasticsearch
// Although this doesn't change the state of the Client, it is a part of its behaviour
func (c *Client) Record(ctx context.Context, t time.Time, kv []DataType) error {
    payload := getQueryString(t, kv)
    _, err := c.client.Index().
        Index(c.indexName).
        Type(c.typeName).
        BodyString(payload).
        Do(ctx)
    if err != nil {
        return err
    }
    return ctx.Err()
}

// Start begins pulling the data and record them.
// when ctx is canceled, all goroutines will stop what they do.
func (c *Client) Start(ctx context.Context) {
    c.ctx = ctx // This is recoreded here because the client hasn't seen it yet

    jobCh := make(chan context.Context)
    resultCh := make(chan jobResult)
    done := fetch(c.logger, c.target, jobCh, resultCh)
    ticker := time.NewTicker(c.interval)
    for {
        select {
        case <-ticker.C:
            ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
            time.AfterFunc(c.timeout, cancel)
            jobCh <- ctx
        case r := <-resultCh:
            if r.err != nil {
                c.logger.Errorf("%s", r.err)
                continue
            }
            values, err := getValues(r.res)
            if err != nil {
                // TODO: change the context
                c.logger.Errorf("decoding json body: %s", err)
                r.res.Close()
                continue
            }

            err = c.Record(c.ctx, r.time, values)
            if err != nil {
                c.logger.Errorf("recording results: %s", err)
                r.res.Close()
                continue
            }
            r.res.Close()

        case <-c.ctx.Done():
            close(jobCh)
            <-done
            return
        }
    }
}

// Stop begins pulling the data and record them
func (c *Client) Stop() error {
    return nil
}

// TODO: Use JSON encoder instead
func getQueryString(time time.Time, kv []DataType) string {
    timestamp := fmt.Sprintf(`"@timestamp":"%s"`, time.Format("2006-01-02T15:04:05.999999-07:00"))
    l := make([]string, len(kv)+1)
    l[0] = timestamp

    for i, v := range kv {
        l[i+1] = v.String()
    }
    return fmt.Sprintf("{%s}", strings.Join(l, ","))
}

func getValues(r io.Reader) ([]DataType, error) {
    var target map[string]interface{}
    err := json.NewDecoder(r).Decode(&target)
    if err != nil {
        return nil, err
    }

    return convertToActual("", target), nil
}

func convertToActual(prefix string, target map[string]interface{}) (result []DataType) {
    for key, value := range target {
        switch v := value.(type) {
        case map[string]interface{}:
            // we have nested values
            result = append(result, convertToActual(prefix+key+".", v)...)
        default:
            result = append(result, getDataType(prefix+key, v)...)
        }
    }
    return
}
