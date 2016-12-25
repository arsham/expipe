// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/recorder"
)

// This file contains the operation section of the engine and its event loop.

// Start begins pulling the data from DataReaders and chips them to the DataRecorder.
// When the context is cancelled or timed out, the engine closes all job channels
// and sends them a stop signal.
func (e *Engine) Start() {
	var wg sync.WaitGroup
	wg.Add(1)
	e.log.Infof("starting with %d readers", len(e.readers))

	go func() {
		for {
			numGoroutines.Set(int64(runtime.NumGoroutine()))
			time.Sleep(50 * time.Millisecond)
		}
	}()

	stop := make(communication.StopChannel)
	e.recorder.Start(e.ctx, stop)
	e.log.Debugf("started recorder: %s", e.recorder.Name())

	e.startReaders(e.ctx)
	go e.eventLoop()

	select {
	case <-e.ctx.Done():
		done := make(chan struct{})
		select {
		case stop <- done:
			<-done
			e.log.Debugf(recorderGone)
		case <-time.After(5 * time.Second):
			e.log.Warnf("recorder %s didn't stop in time", e.recorder.Name())
		}

		e.log.Debug(contextCanceled)
		wg.Done()
	}
	wg.Wait()
}

func (e *Engine) startReaders(ctx context.Context) {
	e.redmu.RLock()
	readers := e.readers
	e.redmu.RUnlock()

	for red, stop := range readers {
		expReaders.Add(1)

		go func(red reader.DataReader, stop communication.StopChannel) {
			ticker := time.NewTicker(red.Interval())
			red.Start(ctx, stop)
			e.log.Debugf("started reader: %s", red.Name())

		LOOP:
			for {
				select {
				case <-ticker.C:
					// [1] job's life cycle starts here...
					e.log.Debugf("issuing job to: %s", red.Name())
					go e.issueReaderJob(red)

				case <-ctx.Done():
					e.stop()
					e.log.Debug("context has been cancelled, end of startReaders method")
					break LOOP
				}
			}
		}(red, stop)
	}
}

// Reads from DataReaders and issues jobs to DataRecorder.
func (e *Engine) eventLoop() {
	e.log.Info("starting event loop")
	for {
		select {
		case r := <-e.readerResChan:
			// [2] then the result from the readers arrives here.
			e.log.Debugf("received job %s", r.ID)
			go e.redirectToRecorder(r)

		case err := <-e.errorChan:
			// [3] some of readers or the recorder had an error
			e.log.WithField("ID", err.ID).WithField("name", err.Name).Error(err)

		case <-e.ctx.Done():
			e.log.Debug("context closed while in eventLoop", e.ctx.Err().Error())
			e.stop()
			return
		}
	}
}

func (e *Engine) issueReaderJob(red reader.DataReader) {
	readJobs.Add(1)
	// to make sure the reader is behaving.
	timeout := red.Timeout() + time.Duration(10*time.Second)
	timer := time.NewTimer(timeout)

	select {
	case red.JobChan() <- communication.NewReadJob(e.ctx):
		// job was sent, we are done here.
		if !timer.Stop() {
			<-timer.C
		}
		return

	case <-timer.C:
		erroredJobs.Add(1)
		e.log.Warn("time out before job was read")

	case <-e.ctx.Done():
		erroredJobs.Add(1)
		e.log.Warn("main context closed before job was read", e.ctx.Err().Error())
		e.stop()

	}
}

// Be aware that I am closing the stream.
func (e *Engine) redirectToRecorder(result *reader.ReadJobResult) {
	defer result.Res.Close()

	payload := datatype.JobResultDataTypes(result.Res, result.Mapper)
	if payload.Error() != nil {
		erroredJobs.Add(1)
		e.log.Warnf("error in payload", payload.Error())
		return
	}
	recordJobs.Add(1)

	timeout := e.recorder.Timeout() + time.Duration(10*time.Second)
	timer := time.NewTimer(timeout)
	recPayload := &recorder.RecordJob{
		ID:        result.ID,
		Ctx:       e.ctx,
		Payload:   payload,
		IndexName: e.recorder.IndexName(),
		TypeName:  result.TypeName,
		Time:      result.Time,
	}

	// sending payload
	select {
	case e.recorder.PayloadChan() <- recPayload:
		// [4] job was sent
		if !timer.Stop() {
			<-timer.C
		}
		e.log.WithField("ID", result.ID).Debug("payload has been delivered")

	case <-timer.C:
		e.log.Warn("timed-out before receiving the error")

	case <-e.ctx.Done():
		e.log.WithField("ID", result.ID).Warn("main context was closed before receiving the error response", e.ctx.Err().Error())
		if !timer.Stop() {
			<-timer.C
		}
		e.stop()
	}
}

// [5] stop closes the job channels
func (e *Engine) stop() {
	// TODO: wait for the recorders to finish their jobs.
	e.log.Debug("STOP method has been called")
	e.shutdown.Do(func() {
		var wg sync.WaitGroup
		e.redmu.RLock()
		readers := e.readers
		e.redmu.RUnlock()

		for red, stop := range readers {
			wg.Add(1)
			go func(red reader.DataReader, stop communication.StopChannel) {
				done := make(chan struct{})
				select {
				case stop <- done:
					<-done
					e.log.Debugf("reader %s has stopped", red.Name())

				case <-time.After(5 * time.Second):
					e.log.Warnf("reader %s didn't stop in time", red.Name())
				}
				wg.Done()
			}(red, stop)
		}
		wg.Wait()
	})
}
