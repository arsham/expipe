// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "net/http"
    "net/http/httptest"
    "reflect"
    "testing"
    "time"

    "github.com/arsham/expvastic"
    "github.com/arsham/expvastic/lib"
)

type MockExpvarReader struct {
    jobCh    chan context.Context
    resultCh chan expvastic.JobResult
    done     chan struct{}
}

func NewMockExpvarReader(jobs chan context.Context, resCh chan expvastic.JobResult, f func(context.Context)) *MockExpvarReader {
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

func (m *MockExpvarReader) JobChan() chan context.Context        { return m.jobCh }
func (m *MockExpvarReader) ResultChan() chan expvastic.JobResult { return m.resultCh }

func TestExpvarReaderErrors(t *testing.T) {
    log := lib.DiscardLogger()
    ctxReader := NewMockCtxReader("nowhere")
    ctxReader.ContextReadFunc = func(ctx context.Context) (*http.Response, error) {
        return nil, fmt.Errorf("Error")
    }
    reader, _ := expvastic.NewExpvarReader(log, ctxReader)
    reader.Start()
    ctx, cancel := context.WithCancel(context.Background())
    reader.JobChan() <- ctx
    res := <-reader.ResultChan()
    if res.Err == nil || reflect.TypeOf(res.Res) != reflect.TypeOf(new(lib.DummyReadCloser)) {
        t.Errorf("Expecting error and empty results: err (%s), res (%v)", res.Err, res.Res)
    }
    cancel()
}

func TestExpvarReaderReads(t *testing.T) {
    log := lib.DiscardLogger()
    testCase := `{"hey": 666}`
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        io.WriteString(w, testCase)
    }))
    ctxReader := NewMockCtxReader(ts.URL)
    reader, _ := expvastic.NewExpvarReader(log, ctxReader)
    reader.Start()
    ctx, cancel := context.WithCancel(context.Background())
    reader.JobChan() <- ctx
    res := <-reader.ResultChan()
    if res.Err != nil {
        t.Errorf("Expecting no errors, got (%s)", res.Err)
    }
    buf := new(bytes.Buffer)
    buf.ReadFrom(res.Res)
    if buf.String() != testCase {
        t.Errorf("want (%s), got (%s)", testCase, buf.String())
    }

    cancel()
}

func TestExpvarReaderClosesStream(t *testing.T) {
    log := lib.DiscardLogger()
    ctxReader := NewMockCtxReader("nowhere")
    reader, _ := expvastic.NewExpvarReader(log, ctxReader)
    done := reader.Start()
    ctx, cancel := context.WithCancel(context.Background())
    reader.JobChan() <- ctx
    <-reader.ResultChan()
    close(reader.JobChan())
    select {
    case <-done:
    case <-time.After(1 * time.Second):
        t.Error("The channel was not closed in time")
    }
    cancel()
}
