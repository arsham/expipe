// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
	"context"
	"net/http"
	"time"

	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/reader/expvar"
	"github.com/arsham/expvastic/recorder"
)

// Use this setup to test a recorder's behaviour
func simpleRecorderSetup(url string, readerChan chan struct{}, recJobChan chan *recorder.RecordJob) *expvastic.Conf {
	log := lib.DiscardLogger()
	read := &reader.MockCtxReader{
		ContextReadFunc: func(ctx context.Context) (*http.Response, error) {
			readerChan <- struct{}{}
			return http.Get(url)
		},
	}
	rdr, _ := expvar.NewExpvarReader(log, read)
	rdr.Start()

	rec := &recorder.MockRecorder{
		PayloadChanFunc: func() chan *recorder.RecordJob {
			if recJobChan == nil {
				recJobChan = make(chan *recorder.RecordJob)
			}
			return recJobChan
		},
	}

	return &expvastic.Conf{
		Logger:       log,
		TargetReader: rdr,
		Recorder:     rec,
		Target:       url,
		Interval:     1 * time.Millisecond, Timeout: 3 * time.Millisecond,
	}
}

// // Use this setup to test a recorder's behaviour and mock the reader
func simpleRecReaderSetup(url string, readerChan chan struct{}, recJobChan chan *recorder.RecordJob, resultChan chan *reader.ReadJobResult) *expvastic.Conf {
	log := lib.DiscardLogger()
	jobs := make(chan context.Context)

	rdr := reader.NewMockExpvarReader(jobs, resultChan, func(c context.Context) {})

	rec := &recorder.MockRecorder{
		PayloadChanFunc: func() chan *recorder.RecordJob {
			return recJobChan
		},
	}

	return &expvastic.Conf{
		Logger:       log,
		TargetReader: rdr,
		Recorder:     rec,
		Target:       url,
		Interval:     1 * time.Millisecond, Timeout: 3 * time.Millisecond,
	}
}
