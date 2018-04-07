// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine

import (
	"context"
	"runtime"
	"time"

	"github.com/arsham/expipe/tools"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/recorder"
	"github.com/arsham/expipe/tools/token"
	"github.com/pkg/errors"
)

var chanBuffer = 100

// Start begins pulling data from DataReader and chip them to the DataRecorder.
// When the context is cancelled or timed out, the engine abandons its
// operations and returns an error if accrued.
func Start(e Engine) chan struct{} {
	stop := make(chan struct{})
	go func() {
		dispatch := dispatchLoop(e.Ctx(), e.Log(), e.Recorders())
		for {
			if ok := iterate(e, dispatch, stop); !ok {
				return
			}
		}
	}()
	go func() {
		for {
			numGoroutines.Set(int64(runtime.NumGoroutine()))
			time.Sleep(50 * time.Millisecond)
		}
	}()
	return stop
}

func iterate(e Engine, dispatch chan *reader.Result, stop chan struct{}) bool {
	timer := time.NewTimer(e.Reader().Interval())
	select {
	case <-timer.C:
		waitingReadJobs.Add(1)
		defer waitingReadJobs.Add(-1)
		job := token.New(e.Ctx())
		res, err := e.Reader().Read(job)
		if errors.Cause(err) != nil {
			erroredJobs.Add(1)
			e.Log().Errorf("read job: %v", err)
			break
		}
		if res == nil || res.Content == nil {
			erroredJobs.Add(1)
			e.Log().Errorf("read job: %v", err)
			break
		}
		readJobs.Add(1)
		dispatch <- res
	case <-e.Ctx().Done():
		close(stop)
		return false
	}
	return true
}

// dispatchLoop starts a goroutine for each recorder and fans out the results.
// Engine can send send the results through the returning channel.
func dispatchLoop(ctx context.Context, log tools.FieldLogger, recs map[string]recorder.DataRecorder) chan *reader.Result {
	dispatch := make(chan *reader.Result, len(recs)*chanBuffer)
	ring := make([]chan *reader.Result, len(recs))
	var i int
	for _, rec := range recs {
		d := make(chan *reader.Result, chanBuffer)
		ring[i] = d
		i++
		go dispatchRecord(ctx, log, rec, d)
	}
	go fanOut(ring, dispatch)
	return dispatch
}

func dispatchRecord(ctx context.Context, log tools.FieldLogger, rec recorder.DataRecorder, dispatch chan *reader.Result) {
	for {
		select {
		case result := <-dispatch:
			res := make([]byte, len(result.Content))
			copy(res, result.Content)
			payload, err := datatype.JobResultDataTypes(res, result.Mapper.Copy())
			if err != nil {
				log.Errorf("error in payload: %s", err)
				return
			}
			waitingRecordJobs.Add(1)
			job := recorder.Job{
				ID:        result.ID,
				Payload:   payload,
				IndexName: rec.IndexName(),
				TypeName:  result.TypeName,
				Time:      result.Time,
			}
			err = rec.Record(ctx, job)
			if err != nil {
				waitingRecordJobs.Add(-1)
				log.Errorf("record error: %v", err)
			}
			waitingRecordJobs.Add(-1)
			recordJobs.Add(1)
		case <-ctx.Done():
		}
	}
}

// fanOut sends each job from dispatch to all ring channels.
// It starts a goroutine for each job.
func fanOut(ring []chan *reader.Result, dispatch chan *reader.Result) {
	for {
		job := <-dispatch
		for _, r := range ring {
			go func(r chan *reader.Result) {
				r <- job
			}(r)
		}
	}
}
