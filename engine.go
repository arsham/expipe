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

// Engine represents an engine that receives information from readers and ships them to recorders
// The Engine is allowed to change the index and type names at will
type Engine struct {
	ctx          context.Context // When this context is canceled, it client tries to finalize its work
	targetReader TargetReader    // The worker that reads from an expvar provider
	recorder     DataRecorder    // Recorder (e.g. ElasticSearch) client
	indexName    string          // Recorder (e.g. ElasticSearch) index name
	typeName     string          // Recorder (e.g. ElasticSearch) type name
	interval     time.Duration
	timeout      time.Duration
	logger       logrus.FieldLogger
}

// NewEngine creates an index if not exists
// It returns an error if index creation is unsuccessful
func NewEngine(ctx context.Context, c Conf) *Engine {
	cl := &Engine{
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
func (c *Engine) Start() {
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
				errCh := make(chan error)
				p := c.recorder.PayloadChan()
				payload := &RecordJob{
					Ctx:       c.ctx,
					Payload:   jobResultDataTypes(r.Res),
					IndexName: c.indexName,
					TypeName:  c.typeName,
					Time:      r.Time,
					Err:       errCh,
				}
				p <- payload
				if err := <-errCh; err != nil {
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
func (c *Engine) Stop() error {
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
