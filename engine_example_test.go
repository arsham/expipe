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
    "time"

    "github.com/arsham/expvastic"
    "github.com/arsham/expvastic/lib"
    "github.com/arsham/expvastic/reader"
    "github.com/arsham/expvastic/recorder"
)

func ExampleEngine_sendJob() {
    var res *reader.ReadJobResult
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())
    desire := `{"the key": "is the value!"}`

    redTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        io.WriteString(w, desire)
    }))
    defer redTs.Close()

    recTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("Job was recorded")
    }))
    defer recTs.Close()

    ctxReader := reader.NewCtxReader(redTs.URL)
    red, _ := reader.NewSimpleReader(log, ctxReader, "reader_example", 10*time.Millisecond, 10*time.Millisecond)
    rec, _ := recorder.NewSimpleRecorder(ctx, log, "reader_example", recTs.URL, "intexName", "typeName", 10*time.Millisecond, 10*time.Millisecond)
    redDone := red.Start(ctx)
    recDone := rec.Start(ctx)

    cl, err := expvastic.NewWithReadRecorder(ctx, log, red, rec)
    fmt.Println("Engine creation success:", err == nil)
    clDone := cl.Start()

    select {
    case red.JobChan() <- ctx:
        fmt.Println("Just sent a job request")
    case <-time.After(time.Second):
        panic("expected the reader to recive the job, but it blocked")
    }

    select {
    case res = <-red.ResultChan():
        fmt.Println("Job operation success:", res.Err == nil)
    case <-time.After(5 * time.Second): // Should be more than the interval, otherwise the response is not ready yet
        panic("expected to recive a data back, nothing recieved")
    }

    buf := new(bytes.Buffer)
    buf.ReadFrom(res.Res)
    fmt.Println("Reader just received payload:", buf.String())

    cancel()

    _, open := <-redDone
    fmt.Println("Reader closure:", !open)

    _, open = <-recDone
    fmt.Println("Recorder closure:", !open)

    _, open = <-clDone
    fmt.Println("Client closure:", !open)

    // Output:
    // Engine creation success: true
    // Just sent a job request
    // Job was recorded
    // Job operation success: true
    // Reader just received payload: {"the key": "is the value!"}
    // Reader closure: true
    // Recorder closure: true
    // Client closure: true
}
