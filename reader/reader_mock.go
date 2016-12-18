// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import "context"

type MockExpvarReader struct {
	jobCh     chan context.Context
	resultCh  chan *ReadJobResult
	done      chan struct{}
	StartFunc func() chan struct{}
}

func NewMockExpvarReader(jobs chan context.Context, resCh chan *ReadJobResult, f func(context.Context)) *MockExpvarReader {
	w := &MockExpvarReader{
		jobCh:    jobs,
		resultCh: resCh,
		done:     make(chan struct{}),
	}
	go func() {
		for job := range w.jobCh {
			f(job)
		}
		close(w.done)
	}()
	return w
}

func (m *MockExpvarReader) JobChan() chan context.Context   { return m.jobCh }
func (m *MockExpvarReader) ResultChan() chan *ReadJobResult { return m.resultCh }

func (m *MockExpvarReader) Start() chan struct{} {
	if m.StartFunc != nil {
		return m.StartFunc()
	}
	return nil
}
