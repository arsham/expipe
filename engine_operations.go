// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expipe

import (
	"runtime"
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/recorder"
	"github.com/arsham/expipe/token"
	"github.com/pkg/errors"
)

// This file contains the operation section of the engine and its event loop.

// Start begins pulling the data from DataReaders and chips them to the DataRecorder.
// When the context is cancelled or timed out, the engine abandons its operations.
func (e *Engine) Start() {
	e.log.Infof("starting with %d readers", len(e.readers))
	e.shutdown = make(chan struct{})

	go func() {
		for {
			numGoroutines.Set(int64(runtime.NumGoroutine()))
			time.Sleep(50 * time.Millisecond)
		}
	}()

	e.redmu.RLock()
	// TODO: if "self" is the only reader, quit.
	for _, red := range e.readers {
		e.wg.Add(1)
		go e.readerEventLoop(red)
	}
	e.redmu.RUnlock()
	e.wg.Wait()
}

// readerEventLoop starts readers event loop. It handles the recordings
func (e *Engine) readerEventLoop(red reader.DataReader) {
	expReaders.Add(1)
	ticker := time.NewTicker(red.Interval())
	e.log.Debugf("starting reader: %s", red.Name())
	// Signals the loop to stop for this reader.
	stop := make(chan struct{})
	errChan := make(chan ErrJob, 1000) // TODO: decide this number.
LOOP:
	for {
		select {
		case <-ticker.C:
			// [1] job's life cycle starts here...
			e.log.Debugf("issuing job to: %s", red.Name())
			waitingReadJobs.Add(1)
			go e.issueReaderJob(red, errChan, stop)
		case job := <-e.readerJobs:
			// note that the job is not necessarily from current reader as they
			// all share the same channel (see below).
			// However it's ok to ship to the recorder as they are ship their
			// results to the same recorder.
			waitingRecordJobs.Add(1)
			go e.shipToRecorder(job, errChan)
		case err := <-errChan:
			e.log.Error(err.Error())
		case <-stop:
			// This reader is quitting.
			break LOOP
		case <-e.shutdown:
			// This engine is quitting and all other readers will stop.
			e.log.Debugf("shutting down the engine %s", e)
			break LOOP
		case <-e.ctx.Done():
			e.log.Debug(contextCanceled)
			break LOOP
		}
	}
	e.removeReader(red) // TEST: test the reader is been removed.
	e.log.Debugf("reader %s is down", red.Name())
	e.wg.Done()
}

func (e *Engine) issueReaderJob(red reader.DataReader, errChan chan ErrJob, stop chan struct{}) {
	defer waitingReadJobs.Add(-1)
	readJobs.Add(1)
	select {
	case <-e.shutdown:
		return //the engine has been already shut down
	default:
	}
	// to make sure the reader is behaving.
	timeout := red.Timeout() + 10*time.Second
	timer := time.NewTimer(timeout)
	done := make(chan struct{})
	job := token.New(e.ctx)
	go func() {
		// TODO: When reader/recorder are not available, don't check right away.
		// Have a lock on the reader so the next ticker wouldn't bother.
		res, err := red.Read(job)
		if err != nil {
			if err == reader.ErrBackoffExceeded {
				close(stop)
			}
			errChan <- ErrJob{
				ID:   job.ID(),
				Name: red.Name(),
				Err:  errors.Wrap(err, "issuing reader job"),
			}
			return
		}
		e.readerJobs <- res
		close(done)
	}()
	select {
	case <-done:
		// job was sent, we are done here.
		if !timer.Stop() {
			<-timer.C
		}
		return
	case <-timer.C:
		erroredJobs.Add(1)
		e.log.Warn("time out before job was read")
	case <-e.ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		erroredJobs.Add(1)
		e.log.Warn("main context closed before job was read", e.ctx.Err().Error())
	}
}

func (e *Engine) shipToRecorder(result *reader.Result, errChan chan ErrJob) {
	defer waitingRecordJobs.Add(-1)
	res := make([]byte, len(result.Content))
	copy(res, result.Content)
	payload, err := datatype.JobResultDataTypes(res, result.Mapper.Copy())
	if err != nil {
		e.log.Warnf("error in payload: %s", err)
		return
	}
	recordJobs.Add(1)
	timeout := e.recorder.Timeout() + 10*time.Second
	timer := time.NewTimer(timeout)
	recPayload := &recorder.Job{
		ID:        result.ID,
		Payload:   payload,
		IndexName: e.recorder.IndexName(),
		TypeName:  result.TypeName,
		Time:      result.Time,
	}
	done := make(chan struct{})
	go func() {
		// sending payload
		err := e.recorder.Record(e.ctx, recPayload)
		if err != nil {
			if err == recorder.ErrBackoffExceeded {
				close(e.shutdown)
			}
			errChan <- ErrJob{
				ID:   result.ID,
				Name: e.recorder.Name(),
				Err:  errors.Wrap(err, "sending payload"),
			}
		}
		close(done)
	}()
	select {
	case <-done:
		// [4] job was sent
		if !timer.Stop() {
			<-timer.C
		}
		e.log.WithField("ID", result.ID).Debug("payload has been delivered")
	case <-timer.C:
		e.log.Debug("timed-out before receiving the error")
	case <-e.ctx.Done():
		e.log.WithField("ID", result.ID).Debug("main context was closed before receiving the error response", e.ctx.Err().Error())
		if !timer.Stop() {
			<-timer.C
		}
	}
}

// removes the reader from the readers map. If the reader is already gone, it
// ignores the action
func (e *Engine) removeReader(r reader.DataReader) {
	e.redmu.Lock()
	delete(e.readers, r.Name())
	e.redmu.Unlock()
}
