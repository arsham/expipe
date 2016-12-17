// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import "github.com/arsham/expvastic"

type mockRecorder struct {
    PayloadChanFunc func() chan *expvastic.RecordJob
    ErrorFunc       func() error
    StartFunc       func() chan struct{}
}

func (m *mockRecorder) PayloadChan() chan *expvastic.RecordJob {
    if m.PayloadChanFunc != nil {
        return m.PayloadChanFunc()
    }
    return nil
}

func (m *mockRecorder) Error() error {
    if m.ErrorFunc != nil {
        return m.ErrorFunc()
    }
    return nil
}

func (m *mockRecorder) Start() chan struct{} {
    if m.StartFunc != nil {
        return m.StartFunc()
    }
    return nil
}
