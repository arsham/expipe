// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "net/http"
    "net/http/httptest"
    "time"

    "github.com/arsham/expvastic/lib"
)

func ExampleSimpleReader() {
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        io.WriteString(w, `{"the key": "is the value!"}`)
    }))
    defer ts.Close()

    jobChan := make(chan context.Context, 10)
    resultChan := make(chan *ReadJobResult, 10)
    ctxReader := NewCtxReader(ts.URL)
    red, _ := NewSimpleReader(log, ctxReader, jobChan, resultChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
    done := red.Start(ctx)

    job, _ := context.WithCancel(ctx)
    // Issueing a job
    red.JobChan() <- job

    // Now waiting for the results
    res := <-red.ResultChan()
    fmt.Println("Error:", res.Err)

    // Let's read what it retreived
    buf := new(bytes.Buffer)
    buf.ReadFrom(res.Res)
    fmt.Println("Result is:", buf.String())

    // The reader should finish gracefully
    cancel()
    <-done
    fmt.Println("Readed has finished")
    // We need to cancel the job now
    fmt.Println("All done!")
    // Output:
    // Error: <nil>
    // Result is: {"the key": "is the value!"}
    // Readed has finished
    // All done!
}

func ExampleSimpleReader_start1() {
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, `{"the key": "is the value!"}`) }))
    defer ts.Close()

    jobChan := make(chan context.Context)
    resultChan := make(chan *ReadJobResult)

    red, _ := NewSimpleReader(log, NewCtxReader(ts.URL), jobChan, resultChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
    done := red.Start(ctx)
    fmt.Println("Reader has started its event loop!")

    select {
    case <-done:
        panic("Reader shouldn't have closed its done channel")
    default:
        fmt.Println("Reader is working!")
    }

    cancel()
    <-done
    fmt.Println("Reader has stopped its event loop!")
    // Output:
    // Reader has started its event loop!
    // Reader is working!
    // Reader has stopped its event loop!
}
