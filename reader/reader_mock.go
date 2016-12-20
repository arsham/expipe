// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"context"
	"time"

	"github.com/Sirupsen/logrus"
)

// SimpleReader is useful for testing purposes.
type SimpleReader struct {
	name       string
	jobChan    chan context.Context
	resultChan chan *ReadJobResult
	ctxReader  ContextReader
	logger     logrus.FieldLogger
	StartFunc  func() chan struct{}
}

func NewSimpleReader(logger logrus.FieldLogger, ctxReader ContextReader, name string) (*SimpleReader, error) {
	w := &SimpleReader{
		name:       name,
		jobChan:    make(chan context.Context),
		resultChan: make(chan *ReadJobResult),
		ctxReader:  ctxReader,
		logger:     logger,
	}
	return w, nil
}

func (m *SimpleReader) Name() string                    { return m.name }
func (m *SimpleReader) JobChan() chan context.Context   { return m.jobChan }
func (m *SimpleReader) ResultChan() chan *ReadJobResult { return m.resultChan }

func (m *SimpleReader) Start() chan struct{} {
	if m.StartFunc != nil {
		return m.StartFunc()
	}
	done := make(chan struct{})
	go func() {
		for job := range m.jobChan {
			resp, _ := m.ctxReader.Get(job)

			res := &ReadJobResult{
				Time: time.Now(),
				Res:  resp.Body,
			}
			m.resultChan <- res
		}
		close(done)
	}()
	return done
}
