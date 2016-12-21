// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
    "context"
    "fmt"
    "net/http"
    "net/http/httptest"
    "time"

    "github.com/arsham/expvastic/lib"
)

func ExampleSimpleRecorder() {
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("I have received the payload!")
    }))
    defer ts.Close()

    rec, _ := NewSimpleRecorder(ctx, log, "reader_example", ts.URL, "intexName", "typeName", 10*time.Millisecond, 10*time.Millisecond)
    done := rec.Start(ctx)
    errChan := make(chan error)
    job := &RecordJob{
        Ctx:       ctx,
        Payload:   nil,
        IndexName: "my index",
        TypeName:  "my type",
        Time:      time.Now(),
        Err:       errChan,
    }
    // Issueing a job
    rec.PayloadChan() <- job
    rec.PayloadChan() <- job

    // Now waiting for the results
    res := <-errChan
    fmt.Println("Error:", res)

    // The recorder should finish gracefully
    cancel()
    <-done
    fmt.Println("Readed has finished")

    // We need to cancel the job now
    fmt.Println("Finished sending!")
    // close(rec.PayloadChan())
    // Output:
    // I have received the payload!
    // Error: <nil>
    // Finished sending!
    // cReaded has finished
}
