// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
    "bytes"
    "context"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/http/httptest"
    "strings"
    "sync"
    "testing"
    "time"

    "github.com/Sirupsen/logrus"
    "github.com/arsham/expvastic"
    "github.com/arsham/expvastic/lib"
)

type targetReader interface {
    JobChan() chan context.Context
    ResultChan() chan expvastic.JobResult
}

func sampleSetup(log *logrus.Logger, url string, reader targetReader, rec *mockRecorder) *expvastic.Conf {
    return &expvastic.Conf{
        Logger:       log,
        TargetReader: reader,
        Recorder:     rec,
        Target:       url,
        Interval:     1 * time.Millisecond, Timeout: 3 * time.Millisecond,
    }
}

func TestClientSendsTheJob(t *testing.T) {
    log := lib.DiscardLogger()
    bg := context.Background()
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
    defer ts.Close()

    rec := &mockRecorder{
        RecordFunc: func(ctx context.Context, typeName string, t time.Time, kv []expvastic.DataType) error { return nil },
    }
    jobSent := make(chan struct{})
    read := &MockCtxReader{
        ContextReadFunc: func(ctx context.Context) (*http.Response, error) {
            jobSent <- struct{}{}
            return http.Get(ts.URL)
        },
    }
    reader, _ := expvastic.NewExpvarReader(log, read)
    reader.Start()
    conf := sampleSetup(log,
        ts.URL,
        reader,
        rec,
    )
    ctx, cancel := context.WithCancel(bg)
    cl := expvastic.NewClient(ctx, *conf)
    go cl.Start()

    select {
    case <-ctx.Done():
        t.Error("job wasn't sent")
    case <-jobSent:
        cancel()
    }

    <-ctx.Done()
}

func TestClientInspectResultsErrors(t *testing.T) {
    log := lib.DiscardLogger()
    bg := context.Background()
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
    called := new(bool)
    *called = false
    defer ts.Close()
    rec := &mockRecorder{
        RecordFunc: func(ctx context.Context, typeName string, t time.Time, kv []expvastic.DataType) error {
            *called = true
            return nil
        },
    }
    resCh := make(chan expvastic.JobResult)
    jobs := make(chan context.Context)
    reader := NewMockExpvarReader(jobs, resCh, func(c context.Context) {
    })
    conf := sampleSetup(log, ts.URL, reader, rec)
    ctx, _ := context.WithTimeout(bg, 5*time.Millisecond)
    cl := expvastic.NewClient(ctx, *conf)
    defer cl.Stop()
    go cl.Start()
    res := resCh
    msg := ioutil.NopCloser(strings.NewReader("bad json"))
    res <- expvastic.JobResult{Res: msg}
    <-ctx.Done()

    if *called {
        t.Error("Shouldn't have called the recorder")
    }
}

func TestClientRecorderReturnsCorrectResult(t *testing.T) {
    var wg sync.WaitGroup
    bg := context.Background()
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
    defer ts.Close()
    ftype := expvastic.FloatType{"test", 666.66}
    ftypeStr := fmt.Sprintf("{%s}", ftype)
    result := new(expvastic.DataType)
    rec := &mockRecorder{
        RecordFunc: func(ctx context.Context, typeName string, t time.Time, kv []expvastic.DataType) error {
            *result = kv[0]
            wg.Done()
            return nil
        },
    }
    resCh := make(chan expvastic.JobResult)
    jobs := make(chan context.Context)
    reader := NewMockExpvarReader(jobs, resCh, func(c context.Context) {})
    log := lib.DiscardLogger()
    conf := sampleSetup(log, ts.URL, reader, rec)
    ctx, _ := context.WithTimeout(bg, 5*time.Millisecond)
    cl := expvastic.NewClient(ctx, *conf)
    defer cl.Stop()
    buf := new(bytes.Buffer)
    log.Out = buf
    go cl.Start()
    res := reader.ResultChan()
    msg := ioutil.NopCloser(strings.NewReader(ftypeStr))
    wg.Add(1)
    res <- expvastic.JobResult{Res: msg}
    wg.Wait()
    if (*result).String() != ftype.String() {
        t.Errorf("want (%s), got (%s)", ftype.String(), (*result).String())
    }
    <-ctx.Done()
}

func TestClientStop(t *testing.T) {
    t.Skip("Not implemented here")
}
