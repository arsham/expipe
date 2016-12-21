// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
	"context"
	"fmt"
	"strings"
	"sync"
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
// Note that we could create a channel and distribute the recorders payload, but we didn't because there
// is no way to find out which recorder errors right after the payload has been sent.
type Engine struct {
	once         sync.Once
	name         string                           // Name identifier for this engine.
	ctx          context.Context                  // Will call Stop() when this context is canceled/timedout. This is a new context from the parent
	cancel       context.CancelFunc               // Based on the new context
	targetReader reader.TargetReader              // The worker that reads from an expvar provider.
	recorders    map[string]recorder.DataRecorder // Recorder (e.g. ElasticSearch) client. The key is the name of the recorder
	logger       logrus.FieldLogger
}

// NewWithConfig instantiates reader and recorders from the configurations.
func NewWithConfig(ctx context.Context, log logrus.FieldLogger, reader config.ReaderConf, recorders ...config.RecorderConf) (*Engine, error) {
	recs := make([]recorder.DataRecorder, len(recorders))
	for i, recConf := range recorders {
		rec, err := recConf.NewInstance(ctx)
		if err != nil {
			return nil, err
		}
		recs[i] = rec
	}

	red, err := reader.NewInstance(ctx)
	if err != nil {
		return nil, err
	}
	return NewWithReadRecorder(ctx, log, red, recs...)
}

// NewWithReadRecorder creates an instance with already made reader and recorders.
func NewWithReadRecorder(ctx context.Context, log logrus.FieldLogger, red reader.TargetReader, recs ...recorder.DataRecorder) (*Engine, error) {
	recNames := make([]string, len(recs))
	recsMap := make(map[string]recorder.DataRecorder, len(recs))
	for i, rec := range recs {
		recNames[i] = rec.Name()
		recsMap[rec.Name()] = rec
	}
	engineName := fmt.Sprintf("( %s >-x-< %s )", red.Name(), strings.Join(recNames, ","))
	log = log.WithField("engine", engineName)
	cl := &Engine{
		name:         engineName,
		ctx:          ctx,
		recorders:    recsMap,
		targetReader: red,
		logger:       log,
	}
	return cl, nil
}

// Start begins pulling the data from TargetReader and chips them to DataRecorder.
// When the context cancels or timesout, the engine closes all job channels, causing the readers and recorders to stop.
func (e *Engine) Start() chan struct{} {
	done := make(chan struct{})
	go e.once.Do(func() {
		ctx, cancel := context.WithCancel(e.ctx)
		e.cancel = cancel
		readerDone := e.targetReader.Start(ctx)
		for _, rec := range e.recorders {
			// TODO: keep the done channels
			rec.Start(ctx)
		}
		e.readMsgLoop(done, readerDone)
	})
	return done
}

// TODO: test
func (e *Engine) readMsgLoop(selfDone, readerDone chan struct{}) {
	resultChan := e.targetReader.ResultChan()
	ticker := time.NewTicker(e.targetReader.Interval())
	e.logger.Debug("starting")
	for {
		select {
		case <-ticker.C:
			go e.issueReaderJob()
		case r := <-resultChan:
			if r.Err != nil {
				e.logger.Errorf(r.Err.Error())
				continue
			}
			go e.redirectToRecorders(r)
		case <-readerDone:
			e.logger.Debug("reader is gone now")
			e.Stop()
			close(selfDone)
			return
		case <-e.ctx.Done():
			e.logger.Debug("context has been canceled")
			e.Stop()
			close(selfDone)
			return
		}
	}
}

// Name shows the name identifier for this engine
func (e *Engine) Name() string { return e.name }

// Stop closes the job channels
func (e *Engine) Stop() {
	// TODO: should I close the readers/recorders channels here?
	e.logger.Debug("stopping")
	e.cancel()
}

func (e *Engine) issueReaderJob() {
	ctx, _ := context.WithTimeout(e.ctx, e.targetReader.Timeout()) // QUESTION: do I need this?
	timer := time.NewTimer(e.targetReader.Timeout())
	select {
	case e.targetReader.JobChan() <- ctx:
		timer.Stop()
		return
	case <-timer.C: // QUESTION: Do I need this? Or should I apply the same for recorder?
		e.logger.Warn("timedout before receiving the error")
	case <-ctx.Done():
		e.logger.Warnf("timedout before receiving the error response: %s", ctx.Err().Error())
	}

}

// Be aware that I am closing the stream.
func (e *Engine) redirectToRecorders(r *reader.ReadJobResult) {
	defer r.Res.Close()
	payload := datatype.JobResultDataTypes(r.Res)
	for name, rec := range e.recorders {
		go func(name string, rec recorder.DataRecorder) {
			e.logger.Debug("sending payload")
			errChan := make(chan error)
			ctx, _ := context.WithTimeout(e.ctx, rec.Timeout())
			payload := &recorder.RecordJob{
				Ctx:       ctx,
				Payload:   payload,
				IndexName: rec.IndexName(),
				TypeName:  rec.TypeName(),
				Time:      r.Time,
				Err:       errChan,
			}
			rec.PayloadChan() <- payload
			select {
			case err := <-errChan:
				if err != nil {
					e.logger.Errorf("%s", err.Error())
				}
			case <-ctx.Done():
				e.logger.Warnf("timedout before receiving the error: %s", ctx.Err().Error())
			case <-e.ctx.Done():
				e.logger.Warnf("main context was canceled before receiving the error: %s", ctx.Err().Error())
			}
		}(name, rec)
	}
}
