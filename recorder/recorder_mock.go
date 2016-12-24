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

type SimpleRecorder struct {
    name            string
    endpoint        string
    indexName       string
    jobChan         chan *RecordJob
    errorChan       chan<- communication.ErrorMessage
    logger          logrus.FieldLogger
    timeout         time.Duration
    Pmu             sync.RWMutex
    PayloadChanFunc func() chan *RecordJob
    ErrorFunc       func() error
    Smu             sync.RWMutex
    StartFunc       func() chan struct{}
}

func NewSimpleRecorder(ctx context.Context, logger logrus.FieldLogger, payloadChan chan *RecordJob, errorChan chan<- communication.ErrorMessage, name, endpoint, indexName string, timeout time.Duration) (*SimpleRecorder, error) {
    w := &SimpleRecorder{
        name:      name,
        endpoint:  endpoint,
        indexName: indexName,
        jobChan:   payloadChan,
        errorChan: errorChan,
        logger:    logger,
        timeout:   timeout,
    }
    return w, nil
}

func (s *SimpleRecorder) PayloadChan() chan *RecordJob {
    s.Pmu.RLock()
    defer s.Pmu.RUnlock()
    if s.PayloadChanFunc != nil {
        return s.PayloadChanFunc()
    }
    return s.jobChan
}

func (s *SimpleRecorder) Error() error {
    if s.ErrorFunc != nil {
        return s.ErrorFunc()
    }
    return nil
}

func (s *SimpleRecorder) Start(ctx context.Context) <-chan struct{} {
    s.Smu.RLock()
    if s.StartFunc != nil {
        s.Smu.RUnlock()
        return s.StartFunc()
    }
    s.Smu.RUnlock()
    done := make(chan struct{})
    go func() {
    LOOP:
        for {
            select {
            case job := <-s.jobChan:
                go func(job *RecordJob) {
                    res, err := http.Get(s.endpoint)
                    if err != nil {
                        s.errorChan <- communication.ErrorMessage{ID: job.ID, Name: s.Name(), Err: err}
                        return
                    }
                    res.Body.Close()
                }(job)
            case <-ctx.Done():
                break LOOP
            }
        }
        close(done)
    }()
    return done
}

func (s *SimpleRecorder) Name() string           { return s.name }
func (s *SimpleRecorder) IndexName() string      { return s.indexName }
func (s *SimpleRecorder) Timeout() time.Duration { return s.timeout }
