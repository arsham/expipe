// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
	"context"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/config"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/recorder"
)

// Engine represents an engine that receives information from readers and ships them to recorders.
// The Engine is allowed to change the index and type names at will.
// When the context times out or canceled, the engine will close the the job channels by calling the Stop method.
type Engine struct {
	name         string                // Name identifier for this engine.
	ctx          context.Context       // Will call Stop() when this context is canceled/timedout. This is a new context from the parent
	cancel       context.CancelFunc    // Based on the new context
	targetReader reader.TargetReader   // The worker that reads from an expvar provider.
	recorder     recorder.DataRecorder // Recorder (e.g. ElasticSearch) client.
	logger       logrus.FieldLogger
}

// NewWithConfig copies its configurations from c.
func NewWithConfig(ctx context.Context, log logrus.FieldLogger, reader config.ReaderConf, recorder config.RecorderConf) (*Engine, error) {
	rec, err := recorder.NewInstance(ctx)
	if err != nil {
		return nil, err
	}
	red, err := reader.NewInstance(ctx)
	if err != nil {
		return nil, err
	}
	return NewWithReadRecorder(ctx, log, red, rec)
}

// NewWithReadRecorder creates an instance with already made reader and recorder.
func NewWithReadRecorder(ctx context.Context, log logrus.FieldLogger, red reader.TargetReader, rec recorder.DataRecorder) (*Engine, error) {
	ctx, cancel := context.WithCancel(ctx)
	cl := &Engine{
		name:         fmt.Sprintf("( %s >=< %s )", red.Name(), rec.Name()),
		ctx:          ctx,
		cancel:       cancel,
		recorder:     rec,
		targetReader: red,
		logger:       log,
	}
	return cl, nil
}

// Start begins pulling the data from TargetReader and chips them to DataRecorder.
// When the context cancels or timesout, the engine closes all job channels, causing the readers and recorders to stop.
func (c *Engine) Start() chan struct{} {
	done := make(chan struct{})
	go func() {
		c.targetReader.Start()
		c.recorder.Start()
		resultChan := c.targetReader.ResultChan()
		ticker := time.NewTicker(c.targetReader.Interval())
		c.debug("starting")
		for {
			select {
			case <-ticker.C:
				go c.issueReaderJob()
			case r := <-resultChan:
				if r.Err != nil {
					c.error(r.Err.Error())
					continue
				}
				go c.redirectToRecorder(r)
			case <-c.ctx.Done():
				c.debug("context has been canceled")
				close(done)
				c.Stop()
				return
			}
		}
	}()
	return done
}

// Name shows the name identifier for this engine
func (c *Engine) Name() string {
	return c.name
}

// Stop closes the job channels
func (c *Engine) Stop() {
	// TODO: should I close the readers/recorders channels here?
	c.debug("stopping")
	c.cancel()
}

func (c *Engine) issueReaderJob() {
	ctx, _ := context.WithTimeout(c.ctx, c.targetReader.Timeout()) // QUESTION: do I need this?
	timer := time.NewTimer(c.targetReader.Timeout())
	select {
	case c.targetReader.JobChan() <- ctx:
		timer.Stop()
		return
	case <-timer.C: // QUESTION: Do I need this? Or should I apply the same for recorder?
		c.warn("timedout before receiving the error")
	case <-ctx.Done():
		c.warnf("timedout before receiving the error response: %s", ctx.Err().Error())
	}

}

// Be aware that I am closing the stream.
func (c *Engine) redirectToRecorder(r *reader.ReadJobResult) {
	defer r.Res.Close()
	ctx, _ := context.WithTimeout(c.ctx, c.recorder.Timeout()) // QUESTION: do I need this?
	errChan := make(chan error)
	payload := &recorder.RecordJob{
		Ctx:       ctx,
		Payload:   datatype.JobResultDataTypes(r.Res),
		IndexName: c.recorder.IndexName(),
		TypeName:  c.recorder.TypeName(),
		Time:      r.Time,
		Err:       errChan,
	}
	c.recorder.PayloadChan() <- payload
	select {
	case err := <-errChan:
		if err != nil {
			c.errorf("%s", err.Error())
		}
	case <-ctx.Done():
		c.warnf("timedout before receiving the error: %s", ctx.Err().Error())
	}
}

func (c *Engine) debug(msg string)                    { c.logger.Debugf("%s: %s", c.Name(), msg) }
func (c *Engine) debugf(format string, msg ...string) { c.logger.Debugf("%s: "+format, c.Name(), msg) }
func (c *Engine) error(msg string)                    { c.logger.Error("%s: %s", c.Name(), msg) }
func (c *Engine) errorf(format string, msg ...string) { c.logger.Errorf("%s: "+format, c.Name(), msg) }
func (c *Engine) warn(msg string)                     { c.logger.Warn("%s: %s", c.Name(), msg) }
func (c *Engine) warnf(format string, msg ...string)  { c.logger.Warnf("%s: "+format, c.Name(), msg) }
