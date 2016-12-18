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
)

// Use this setup to test a recorder's behaviour
func simpleRecorderSetup(url string, readerChan chan struct{}, recJobChan chan *expvastic.RecordJob) *expvastic.Conf {
	log := lib.DiscardLogger()
	read := &MockCtxReader{
		ContextReadFunc: func(ctx context.Context) (*http.Response, error) {
			readerChan <- struct{}{}
			return http.Get(url)
		},
	}
	reader, _ := expvastic.NewExpvarReader(log, read)
	reader.Start()

	rec := &mockRecorder{
		PayloadChanFunc: func() chan *expvastic.RecordJob {
			if recJobChan == nil {
				recJobChan = make(chan *expvastic.RecordJob)
			}
			return recJobChan
		},
	}

	return &expvastic.Conf{
		Logger:       log,
		TargetReader: reader,
		Recorder:     rec,
		Target:       url,
		Interval:     1 * time.Millisecond, Timeout: 3 * time.Millisecond,
	}
}

// Use this setup to test a recorder's behaviour and mock the reader
func simpleRecReaderSetup(url string, readerChan chan struct{}, recJobChan chan *expvastic.RecordJob, resultChan chan *expvastic.ReadJobResult) *expvastic.Conf {
	log := lib.DiscardLogger()
	jobs := make(chan context.Context)

	reader := NewMockExpvarReader(jobs, resultChan, func(c context.Context) {})

	rec := &mockRecorder{
		PayloadChanFunc: func() chan *expvastic.RecordJob {
			return recJobChan
		},
	}

	return &expvastic.Conf{
		Logger:       log,
		TargetReader: reader,
		Recorder:     rec,
		Target:       url,
		Interval:     1 * time.Millisecond, Timeout: 3 * time.Millisecond,
	}
}
