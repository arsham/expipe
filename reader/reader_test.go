// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
    "bytes"
    "context"
    "io"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/arsham/expvastic/communication"
    "github.com/arsham/expvastic/lib"
)

// The purpose of these tests is to make sure the simple reader, which is a mock,
// works perfect, so other tests can rely on it.
func TestSimpleReaderReceivesJob(t *testing.T) {
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        io.WriteString(w, `{"the key": "is the value!"}`)
    }))
    defer ts.Close()

    jobChan := make(chan context.Context)
    errorChan := make(chan communication.ErrorMessage)
    resultChan := make(chan *ReadJobResult, 1)
    red, _ := NewSimpleReader(log, NewCtxReader(ts.URL), jobChan, resultChan, errorChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
    red.Start(ctx)

    select {
    case red.JobChan() <- communication.NewReadJob(ctx):
    case <-time.After(5 * time.Second):
        t.Error("expected the reader to recive the job, but it blocked")
    }
}

func TestSimpleReaderSendsResult(t *testing.T) {
    var res *ReadJobResult
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    desire := `{"the key": "is the value!"}`

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        io.WriteString(w, desire)
    }))
    defer ts.Close()

    jobChan := make(chan context.Context)
    errorChan := make(chan communication.ErrorMessage)
    resultChan := make(chan *ReadJobResult)
    red, _ := NewSimpleReader(log, NewCtxReader(ts.URL), jobChan, resultChan, errorChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
    red.Start(ctx)

    red.JobChan() <- communication.NewReadJob(ctx)

    select {
    case err := <-errorChan:
        t.Errorf("didn't expect errors, got (%v)", err.Error())
    case <-time.After(20 * time.Millisecond):
    }

    select {
    case res = <-resultChan:
    case <-time.After(5 * time.Second):
        t.Error("expected to recive a data back, nothing recieved")
    }

    buf := new(bytes.Buffer)
    buf.ReadFrom(res.Res)
    if buf.String() != desire {
        t.Errorf("want (%s), got (%s)", desire, buf.String())
    }
}

func TestSimpleReaderReadsOnBufferedChan(t *testing.T) {
    var res *ReadJobResult
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    desire := `{"the key": "is the value!"}`

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        io.WriteString(w, desire)
    }))
    defer ts.Close()

    jobChan := make(chan context.Context, 10)
    errorChan := make(chan communication.ErrorMessage, 10)
    resultChan := make(chan *ReadJobResult)

    red, _ := NewSimpleReader(log, NewCtxReader(ts.URL), jobChan, resultChan, errorChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
    red.Start(ctx)

    red.JobChan() <- communication.NewReadJob(ctx)

    select {
    case err := <-errorChan:
        t.Errorf("didn't expect errors, got (%v)", err.Error())
    case <-time.After(20 * time.Millisecond):
    }

    select {
    case res = <-resultChan:
    case <-time.After(5 * time.Second):
        t.Error("expected to recive a data back, nothing recieved")
    }

    drained := false
    // Go is fast!
    for i := 0; i < 10; i++ {
        if len(red.JobChan()) == 0 {
            drained = true
            break
        }
        time.Sleep(10 * time.Millisecond)

    }
    if !drained {
        t.Errorf("expected to drain the jobChan, got (%d) left", len(red.JobChan()))
    }

    buf := new(bytes.Buffer)
    buf.ReadFrom(res.Res)
    if buf.String() != desire {
        t.Errorf("want (%s), got (%s)", desire, buf.String())
    }
}

func TestSimpleReaderDrainsAfterClosingContext(t *testing.T) {
    var res *ReadJobResult
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())
    desire := `{"the key": "is the value!"}`

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        io.WriteString(w, desire)
    }))
    defer ts.Close()

    jobChan := make(chan context.Context, 10)
    errorChan := make(chan communication.ErrorMessage, 10)
    resultChan := make(chan *ReadJobResult)

    red, _ := NewSimpleReader(log, NewCtxReader(ts.URL), jobChan, resultChan, errorChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
    red.Start(ctx)

    red.JobChan() <- communication.NewReadJob(ctx)

    select {
    case err := <-errorChan:
        t.Errorf("didn't expect errors, got (%v)", err.Error())
    case <-time.After(20 * time.Millisecond):
    }

    select {
    case res = <-resultChan:
    case <-time.After(5 * time.Second):
        t.Error("expected to recive a data back, nothing recieved")
    }

    cancel()
    drained := false
    // Go is fast!
    for i := 0; i < 10; i++ {
        if len(red.JobChan()) == 0 {
            drained = true
            break
        }
        time.Sleep(10 * time.Millisecond)

    }
    if !drained {
        t.Errorf("expected to drain the jobChan, got (%d) left", len(red.JobChan()))
    }

    buf := new(bytes.Buffer)
    buf.ReadFrom(res.Res)
    if buf.String() != desire {
        t.Errorf("want (%s), got (%s)", desire, buf.String())
    }
}

func TestSimpleReaderCloses(t *testing.T) {
    var res *ReadJobResult
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())
    desire := `{"the key": "is the value!"}`

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        io.WriteString(w, desire)
    }))
    defer ts.Close()

    jobChan := make(chan context.Context)
    errorChan := make(chan communication.ErrorMessage)
    resultChan := make(chan *ReadJobResult)
    red, _ := NewSimpleReader(log, NewCtxReader(ts.URL), jobChan, resultChan, errorChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
    done := red.Start(ctx)

    red.JobChan() <- communication.NewReadJob(ctx)
    res = <-resultChan
    defer res.Res.Close()
    cancel()

    select {
    case <-done:
    case <-time.After(5 * time.Second):
        t.Error("expected to be done with the reader, but it blocked")
    }
}

func TestSimpleReaderClosesWithBufferedChans(t *testing.T) {
    var res *ReadJobResult
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())
    desire := `{"the key": "is the value!"}`

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        io.WriteString(w, desire)
    }))
    defer ts.Close()

    jobChan := make(chan context.Context, 1000)
    errorChan := make(chan communication.ErrorMessage, 1000)
    resultChan := make(chan *ReadJobResult, 1000)
    red, _ := NewSimpleReader(log, NewCtxReader(ts.URL), jobChan, resultChan, errorChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
    done := red.Start(ctx)

    red.JobChan() <- communication.NewReadJob(ctx)
    res = <-resultChan
    defer res.Res.Close()

    cancel()
    select {
    case <-done:
    case <-time.After(5 * time.Second):
        t.Error("expected to be done with the reader, but it blocked")
    }
}
