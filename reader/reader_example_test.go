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

    "github.com/arsham/expvastic/lib"
)

func ExampleSimpleReader() {
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        io.WriteString(w, `{"the key": "is the value!"}`)
    }))
    defer ts.Close()

    ctxReader := NewCtxReader(ts.URL)
    rdr, _ := NewSimpleReader(log, ctxReader, "reader_example")
    done := rdr.Start()

    job, _ := context.WithCancel(ctx)
    // Issueing a job
    rdr.JobChan() <- job

    // Now waiting for the results
    res := <-rdr.ResultChan()
    fmt.Println("Error:", res.Err)

    // Let's read what it retreived
    buf := new(bytes.Buffer)
    buf.ReadFrom(res.Res)
    fmt.Println("Result is:", buf.String())

    // The reader should finish gracefully
    go func() {
        <-done
        fmt.Println("Readed has finished")
    }()
    // We need to cancel the job now
    cancel()
    fmt.Println("All done!")

    // Output:
    // Error: <nil>
    // Result is: {"the key": "is the value!"}
    // All done!
}
