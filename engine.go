// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
	"context"
	"expvar"
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

var (
	expRecorders = expvar.NewInt("Recorders")
	expReaders   = expvar.NewInt("Readers")
	readJobs     = expvar.NewInt("Read Jobs")
	recordJobs   = expvar.NewInt("Record Jobs")
	erroredJobs  = expvar.NewInt("Errored Jobs")
)

// Engine represents an engine that receives information from readers and ships them to recorders.
// The Engine is allowed to change the index and type names at will.
// When the context times out or canceled, the engine will close the the job channels by calling the Stop method.
// Note that we could create a channel and distribute the recorders payload, but we didn't because there
// is no way to find out which recorder errors right after the payload has been sent.
// IMPORTANT: the readers should not close their streams, I am closing them here.
type Engine struct {
	once            sync.Once                        // For guarding the Start method
	name            string                           // Name identifier for this engine.
	ctx             context.Context                  // Will call Stop() when this context is canceled/timedout. This is a new context from the parent.
	cancel          context.CancelFunc               // Based on the new context.
	dataReader      reader.DataReader                // Reads from an expvar provider.
	recorders       map[string]recorder.DataRecorder // Records to ElasticSearch client. The key is the name of the recorder.
	recorderResChan <-chan error
	observer        *observer
	logger          logrus.FieldLogger
}

// NewWithConfig instantiates reader and recorders from the configurations.
func NewWithConfig(ctx context.Context, log logrus.FieldLogger, readChanBuff, readResChanBuff, recChanBuff, recResChan int, red config.ReaderConf, recorders ...config.RecorderConf) (*Engine, error) {
	recs := make([]recorder.DataRecorder, len(recorders))
	jobChan := make(chan context.Context, readChanBuff)
	resultChan := make(chan *reader.ReadJobResult, readResChanBuff)
	for i, recConf := range recorders {
		payloadChan := make(chan *recorder.RecordJob, recChanBuff)
		rec, err := recConf.NewInstance(ctx, payloadChan)
		if err != nil {
			return nil, err
		}
		recs[i] = rec
	}
	r, err := red.NewInstance(ctx, jobChan, resultChan)
	if err != nil {
		return nil, err
	}
	return NewWithReadRecorder(ctx, log, recResChan, r, recs...)
}

// NewWithReadRecorder creates an instance with already made reader and recorders.
// It spawns one reader and streams its payload to all recorders.
// Returns an error if there are recorders with the same name, or any of them have no name.
func NewWithReadRecorder(ctx context.Context, logger logrus.FieldLogger, recResChan int, red reader.DataReader, recs ...recorder.DataRecorder) (*Engine, error) {
	recorderResChan := make(chan error, recResChan)
	recNames := make([]string, len(recs))
	recsMap := make(map[string]recorder.DataRecorder, len(recs))
	seenNames := make(map[string]struct{}, len(recs))
	for i, rec := range recs {
		if rec.Name() == "" {
			return nil, ErrEmptyRecName
		}
		if _, ok := seenNames[rec.Name()]; ok {
			return nil, ErrDupRecName
		}
		seenNames[rec.Name()] = struct{}{}
		recNames[i] = rec.Name()
		recsMap[rec.Name()] = rec
	}

	// just to be cute
	engineName := fmt.Sprintf("( %s >-x-< %s )", red.Name(), strings.Join(recNames, ","))
	logger = logger.WithField("engine", engineName)
	cl := &Engine{
		name:            engineName,
		ctx:             ctx,
		recorders:       recsMap,
		dataReader:      red,
		logger:          logger,
		observer:        newobserver(ctx, logger, recorderResChan, len(recs)),
		recorderResChan: recorderResChan,
	}
	return cl, nil
}

// Start begins pulling the data from DataReader and chips them to DataRecorder.
// When the context is canceled or timed out, the engine closes all job channels, causing the readers and recorders to stop.
func (e *Engine) Start() <-chan struct{} {
	done := make(chan struct{})
	e.logger.Debugf("starting with %d recorders", len(e.recorders))

	go e.once.Do(func() {
		ctx, cancel := context.WithCancel(e.ctx)
		e.cancel = cancel
		readerDone := e.dataReader.Start(ctx)
		expReaders.Add(1)
		for _, rec := range e.recorders {
			e.observer.add(ctx, rec)
			expReaders.Add(1)
		}

		e.eventLoop(readerDone)
		close(done)
	})

	return done
}

// TODO: test
// reads from DataReader and issues jobs to DataRecorders.
func (e *Engine) eventLoop(readerDone <-chan struct{}) {
	resultChan := e.dataReader.ResultChan()
	ticker := time.NewTicker(e.dataReader.Interval())
	e.logger.Debug("starting message loop")
	for {
		select {
		case <-ticker.C:
			// job's life cycle starts here...
			go e.issueReaderJob()
		case r := <-resultChan:
			// ...then the result from the job's outcome arrives here.
			if r.Err != nil {
				e.logger.Errorf(r.Err.Error())
				continue
			}
			go e.redirectToRecorders(r)
		case <-readerDone:
			e.logger.Debug("reader is gone now")
			e.Stop()
			return
		case err := <-e.recorderResChan:
			// ... recorder should inform us with the results
			if err != nil {
				e.logger.Errorf("%s", err)
			}
		case <-e.ctx.Done():
			e.logger.Debug("context has been canceled")
			e.Stop()
			return
		}
	}
}

// Name shows the name identifier for this engine
func (e *Engine) Name() string { return e.name }

// Stop closes the job channels
func (e *Engine) Stop() {
	// TODO: wait for the reader/recorders to finish their jobs.
	e.logger.Debug("stopping")
	e.cancel()
}

func (e *Engine) issueReaderJob() {
	readJobs.Add(1)
	// to make sure the reader is behaving.
	timeout := e.dataReader.Timeout() + time.Duration(10*time.Second)
	timer := time.NewTimer(timeout)
	select {
	case e.dataReader.JobChan() <- e.ctx:
		// job was sent, we are done here.
		timer.Stop()
		return
	case <-timer.C:
		erroredJobs.Add(1)
		e.logger.Warn("timedout before job was read")
	case <-e.ctx.Done():
		erroredJobs.Add(1)
		e.logger.Warnf("main context closed before job was read: %s", e.ctx.Err().Error())
	}
}

// Be aware that I am closing the stream.
// TODO: test the error
func (e *Engine) redirectToRecorders(r *reader.ReadJobResult) {
	defer r.Res.Close()
	payload := datatype.JobResultDataTypes(r.Res)
	if payload.Error() != nil {
		erroredJobs.Add(1)
		e.logger.Warnf("error in payload", payload.Error())
		return
	}
	recordJobs.Add(1)
	e.observer.send(e.ctx, r.TypeName, r.Time, payload)
}
