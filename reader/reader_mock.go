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
	logger     logrus.FieldLogger
	interval   time.Duration
	timeout    time.Duration
	StartFunc  func() chan struct{}
}

func NewSimpleReader(
	logger logrus.FieldLogger,
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
		logger:     logger,
		timeout:    timeout,
		interval:   interval,
	}
	return w, nil
}

func (m *SimpleReader) Start(ctx context.Context) <-chan struct{} {
	if m.StartFunc != nil {
		return m.StartFunc()
	}
	done := make(chan struct{})
	go func() {
	LOOP:
		for {
			select {
			case job := <-m.jobChan:
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
				}
				m.resultChan <- res
			case <-ctx.Done():
				break LOOP
			}
		}
		close(done)
	}()
	return done
}

func (m *SimpleReader) Name() string                    { return m.name }
func (m *SimpleReader) TypeName() string                { return m.typeName }
func (m *SimpleReader) Mapper() datatype.Mapper         { return m.mapper }
func (m *SimpleReader) JobChan() chan context.Context   { return m.jobChan }
func (m *SimpleReader) ResultChan() chan *ReadJobResult { return m.resultChan }
func (m *SimpleReader) Interval() time.Duration         { return time.Second }
func (m *SimpleReader) Timeout() time.Duration          { return time.Second }
