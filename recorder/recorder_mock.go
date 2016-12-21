// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
    "context"
    "net/http"
    "time"

    "github.com/Sirupsen/logrus"
)

type SimpleRecorder struct {
    name      string
    endpoint  string
    indexName string
    typeName  string
    jobChan   chan *RecordJob
    logger    logrus.FieldLogger
    interval  time.Duration
    timeout   time.Duration

    PayloadChanFunc func() chan *RecordJob
    ErrorFunc       func() error
    StartFunc       func() chan struct{}
}

func NewSimpleRecorder(ctx context.Context, logger logrus.FieldLogger, name, endpoint, indexName, typeName string, interval, timeout time.Duration) (*SimpleRecorder, error) {
    w := &SimpleRecorder{
        name:      name,
        endpoint:  endpoint,
        indexName: indexName,
        typeName:  typeName,
        jobChan:   make(chan *RecordJob),
        logger:    logger,
        timeout:   timeout,
        interval:  interval,
    }
    return w, nil
}

func (s *SimpleRecorder) PayloadChan() chan *RecordJob {
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

func (s *SimpleRecorder) Start(ctx context.Context) chan struct{} {
    if s.StartFunc != nil {
        return s.StartFunc()
    }
    done := make(chan struct{})
    go func() {
    LOOP:
        for {
            select {
            case job := <-s.jobChan:
                go func(job *RecordJob) {
                    res, err := http.Get(s.endpoint)
                    if err != nil {
                        res.Body.Close()
                    }
                    job.Err <- err
                }(job)
            case <-ctx.Done():
                break LOOP
            }
        }
        close(done)
    }()
    return done
}

func (s *SimpleRecorder) Name() string            { return s.name }
func (s *SimpleRecorder) IndexName() string       { return s.indexName }
func (s *SimpleRecorder) TypeName() string        { return s.typeName }
func (s *SimpleRecorder) Interval() time.Duration { return s.interval }
func (s *SimpleRecorder) Timeout() time.Duration  { return s.timeout }
