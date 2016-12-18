// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

type MockRecorder struct {
    PayloadChanFunc func() chan *RecordJob
    ErrorFunc       func() error
    StartFunc       func() chan struct{}
}

func (m *MockRecorder) PayloadChan() chan *RecordJob {
    if m.PayloadChanFunc != nil {
        return m.PayloadChanFunc()
    }
    return make(chan *RecordJob)
}

func (m *MockRecorder) Error() error {
    if m.ErrorFunc != nil {
        return m.ErrorFunc()
    }
    return nil
}

func (m *MockRecorder) Start() chan struct{} {
    if m.StartFunc != nil {
        return m.StartFunc()
    }
    return nil
}
