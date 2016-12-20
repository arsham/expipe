// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
    "context"
    "fmt"
    "net/http"

    "github.com/Sirupsen/logrus"
)

type SimpleRecorder struct {
    name      string
    endpoint  string
    indexName string
    jobChan   chan *RecordJob
    logger    logrus.FieldLogger

    PayloadChanFunc func() chan *RecordJob
    ErrorFunc       func() error
    StartFunc       func() chan struct{}
}

func NewSimpleRecorder(ctx context.Context, logger logrus.FieldLogger, name, endpoint, indexName string) (*SimpleRecorder, error) {
    w := &SimpleRecorder{
        name:      name,
        endpoint:  endpoint,
        indexName: indexName,
        jobChan:   make(chan *RecordJob),
        logger:    logger,
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

func (s *SimpleRecorder) Start() chan struct{} {
    if s.StartFunc != nil {
        return s.StartFunc()
    }
    done := make(chan struct{})
    go func() {
        for job := range s.jobChan {
            go func(job *RecordJob) {
                res, err := http.Get(s.endpoint)
                if err != nil {
                    res.Body.Close()
                }
                job.Err <- err
            }(job)
        }
        fmt.Println("dine")
        close(done)
    }()
    return done
}

func (s *SimpleRecorder) Name() string { return s.name }
