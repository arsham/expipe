// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "net/http/httptest"
    "time"

    "github.com/arsham/expvastic/communication"
    "github.com/arsham/expvastic/lib"
)

func ExampleSimpleRecorder() {
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("I have received the payload!")
    }))
    defer ts.Close()

    payloadChan := make(chan *RecordJob)
    errorChan := make(chan communication.ErrorMessage)
    rec, _ := NewSimpleRecorder(ctx, log, payloadChan, errorChan, "reader_example", ts.URL, "intexName", 10*time.Millisecond)
    done := rec.Start(ctx)

    job := &RecordJob{
        Ctx:       ctx,
        Payload:   nil,
        IndexName: "my index",
        Time:      time.Now(),
    }
    // Issueing a job
    rec.PayloadChan() <- job

    // Lets check the errors
    select {
    case <-errorChan:
        panic("Wasn't expecting any errors")
    default:
        fmt.Println("No errors reported")
    }

    // Issueing another job
    rec.PayloadChan() <- job

    // The recorder should finish gracefully
    cancel()
    <-done
    fmt.Println("Readed has finished")

    // We need to cancel the job now
    fmt.Println("Finished sending!")
    // close(rec.PayloadChan())
    // Output:
    // I have received the payload!
    // No errors reported
    // Finished sending!
    // cReaded has finished
}

func ExampleSimpleRecorder_start() {
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, `{"the key": "is the value!"}`) }))
    defer ts.Close()

    payloadChan := make(chan *RecordJob)
    errorChan := make(chan communication.ErrorMessage)
    rec, _ := NewSimpleRecorder(ctx, log, payloadChan, errorChan, "reader_example", ts.URL, "intexName", 10*time.Millisecond)
    done := rec.Start(ctx)

    fmt.Println("Recorder has started its event loop!")

    select {
    case <-done:
        panic("Recorder shouldn't have closed its done channel")
    default:
        fmt.Println("Recorder is working!")
    }

    cancel()
    <-done
    fmt.Println("Recorder has stopped its event loop!")
    // Output:
    // Recorder has started its event loop!
    // Recorder is working!
    // Recorder has stopped its event loop!
}
