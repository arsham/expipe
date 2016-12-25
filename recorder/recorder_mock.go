// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/communication"
)

// SimpleRecorder is designed to be used in tests
type SimpleRecorder struct {
	name            string
	endpoint        string
	indexName       string
	payloadChan     chan *RecordJob
	errorChan       chan<- communication.ErrorMessage
	log             logrus.FieldLogger
	timeout         time.Duration
	Pmu             sync.RWMutex
	PayloadChanFunc func() chan *RecordJob
	ErrorFunc       func() error
	Smu             sync.RWMutex
	StartFunc       func(communication.StopChannel)
}

// NewSimpleRecorder returns a SimpleRecorder instance
func NewSimpleRecorder(ctx context.Context, log logrus.FieldLogger, payloadChan chan *RecordJob, errorChan chan<- communication.ErrorMessage, name, endpoint, indexName string, timeout time.Duration) (*SimpleRecorder, error) {
	w := &SimpleRecorder{
		name:        name,
		endpoint:    endpoint,
		indexName:   indexName,
		payloadChan: payloadChan,
		errorChan:   errorChan,
		log:         log,
		timeout:     timeout,
	}
	return w, nil
}

// PayloadChan returns the payload channel
func (s *SimpleRecorder) PayloadChan() chan *RecordJob {
	s.Pmu.RLock()
	defer s.Pmu.RUnlock()
	if s.PayloadChanFunc != nil {
		return s.PayloadChanFunc()
	}
	return s.payloadChan
}

func (s *SimpleRecorder) Error() error {
	if s.ErrorFunc != nil {
		return s.ErrorFunc()
	}
	return nil
}

// Start calls the StartFunc if exists, otherwise continues as normal
func (s *SimpleRecorder) Start(ctx context.Context, stop communication.StopChannel) {
	s.Smu.RLock()
	if s.StartFunc != nil {
		s.Smu.RUnlock()
		s.StartFunc(stop)
		return
	}
	s.Smu.RUnlock()
	go func() {
		for {
			select {
			case job := <-s.payloadChan:
				go func(job *RecordJob) {
					res, err := http.Get(s.endpoint)
					if err != nil {
						s.errorChan <- communication.ErrorMessage{ID: job.ID, Name: s.Name(), Err: err}
						return
					}
					res.Body.Close()
				}(job)
			case s := <-stop:
				s <- struct{}{}
				return
			}
		}

	}()

}

// Name returns the name
func (s *SimpleRecorder) Name() string { return s.name }

// IndexName returns the indexname
func (s *SimpleRecorder) IndexName() string { return s.indexName }

// Timeout returns the timeout
func (s *SimpleRecorder) Timeout() time.Duration { return s.timeout }
