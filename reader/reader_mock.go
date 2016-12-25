// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"context"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
)

// SimpleReader is useful for testing purposes.
type SimpleReader struct {
	name       string
	typeName   string
	mapper     datatype.Mapper
	jobChan    chan context.Context
	resultChan chan *ReadJobResult
	errorChan  chan<- communication.ErrorMessage
	ctxReader  ContextReader
	log        logrus.FieldLogger
	interval   time.Duration
	timeout    time.Duration
	StartFunc  func(communication.StopChannel)
}

// NewSimpleReader is a reader for using in tests
func NewSimpleReader(
	log logrus.FieldLogger,
	ctxReader ContextReader,
	jobChan chan context.Context,
	resultChan chan *ReadJobResult,
	errorChan chan<- communication.ErrorMessage,
	name,
	typeName string,
	interval,
	timeout time.Duration,
) (*SimpleReader, error) {
	w := &SimpleReader{
		name:       name,
		typeName:   typeName,
		mapper:     &datatype.MapConvertMock{},
		jobChan:    jobChan,
		errorChan:  errorChan,
		resultChan: resultChan,
		ctxReader:  ctxReader,
		log:        log,
		timeout:    timeout,
		interval:   interval,
	}
	return w, nil
}

// Start executes the StartFunc if defined, otherwise continues normally
func (m *SimpleReader) Start(ctx context.Context, stop communication.StopChannel) {
	if m.StartFunc != nil {
		m.StartFunc(stop)
		return
	}
	go func() {
		for {
			select {
			case job := <-m.JobChan():
				id := communication.JobValue(job)
				resp, err := m.ctxReader.Get(job)
				if err != nil {
					m.errorChan <- communication.ErrorMessage{ID: id, Err: err}
					continue
				}
				res := &ReadJobResult{
					ID:       id,
					Time:     time.Now(),
					Res:      resp.Body,
					TypeName: m.TypeName(),
					Mapper:   m.Mapper(),
				}

				m.resultChan <- res
			case s := <-stop:
				s <- struct{}{}
				return
			}
		}
	}()
}

// Name returns the name
func (m *SimpleReader) Name() string { return m.name }

// TypeName returns the typename
func (m *SimpleReader) TypeName() string { return m.typeName }

// Mapper returns the mapper
func (m *SimpleReader) Mapper() datatype.Mapper { return m.mapper }

// JobChan returns the jobchan
func (m *SimpleReader) JobChan() chan context.Context { return m.jobChan }

// ResultChan returns the resultchan
func (m *SimpleReader) ResultChan() chan *ReadJobResult { return m.resultChan }

// Interval returns the interval
func (m *SimpleReader) Interval() time.Duration { return m.interval }

// Timeout returns the timeout
func (m *SimpleReader) Timeout() time.Duration { return m.timeout }

// ErrorChan returns the errorchan
func (m *SimpleReader) ErrorChan() chan<- communication.ErrorMessage { return m.errorChan }
