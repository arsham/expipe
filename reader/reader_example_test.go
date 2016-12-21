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

    ctxReader := NewCtxReader(ts.URL)
    red, _ := NewSimpleReader(log, ctxReader, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
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
