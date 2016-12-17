// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

// Client represents a client that can put information into an ES index
// The Client is allowed to change the index and type names at will
type Client struct {
	ctx          context.Context // When this context is canceled, it client tries to finalize its work
	recorder     Recorder        // ElasticSearch client
	indexName    string          // ElasticSearch index name
	typeName     string          // ElasticSearch type name
	targetReader targetReader    // The worker that reads from an expvar provider
	interval     time.Duration
	timeout      time.Duration
	logger       logrus.FieldLogger
}

// NewClient creates an index if not exists
// It returns an error if index creation is unsuccessful
func NewClient(ctx context.Context, c Conf) *Client {
	cl := &Client{
		ctx:          ctx,
		recorder:     c.Recorder,
		targetReader: c.TargetReader,
		indexName:    c.IndexName,
		typeName:     c.TypeName,
		interval:     c.Interval,
		timeout:      c.Timeout,
		logger:       c.Logger,
	}
	return cl
}

// Start begins pulling the data and record them.
// when ctx is canceled, all goroutines will stop what they do.
func (c *Client) Start() {
	jobChan := c.targetReader.JobChan()
	resultChan := c.targetReader.ResultChan()
	ticker := time.NewTicker(c.interval)
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
			time.AfterFunc(c.timeout, cancel)
			// Issuing the next job
			jobChan <- ctx
		case r := <-resultChan:
			go func() {
				defer r.Res.Close()
				err := c.recorder.Record(c.ctx, c.typeName, r.Time, jobResultDataTypes(r.Res))
				if err != nil {
					c.logger.Errorf("%s", err)
				}
			}()
		case <-c.ctx.Done():
			close(jobChan)
			return
		}
	}
}

// Stop begins pulling the data and record them
func (c *Client) Stop() error {
	return nil
}

// TODO: Use JSON encoder instead
func getQueryString(timestamp time.Time, kv []DataType) string {
	ts := fmt.Sprintf(`"@timestamp":"%s"`, timestamp.Format("2006-01-02T15:04:05.999999-07:00"))
	l := make([]string, len(kv)+1)
	l[0] = ts

	for i, v := range kv {
		l[i+1] = v.String()
	}
	return fmt.Sprintf("{%s}", strings.Join(l, ","))
}
