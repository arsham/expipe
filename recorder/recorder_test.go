// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
    "context"
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/arsham/expvastic/lib"
)

func TestSimpleRecorder(t *testing.T) {
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("I have received the payload!")
    }))
    defer ts.Close()

    rec, _ := NewSimpleRecorder(ctx, log, "reader_example", ts.URL, "intexName", 10*time.Millisecond, 10*time.Millisecond)
    done := rec.Start(ctx)
    errChan := make(chan error)
    job := &RecordJob{
        Ctx:       ctx,
        Payload:   nil,
        IndexName: "my index",
        Time:      time.Now(),
        Err:       errChan,
    }

    select {
    case rec.PayloadChan() <- job:
    case <-time.After(5 * time.Second):
        t.Error("expected the recorder to recive the payload, but it blocked")
    }

    select {
    case err := <-errChan:
        if err != nil {
            t.Errorf("want (nil), got (%v)", err)
        }
    case <-time.After(5 * time.Second):
        t.Error("expected to recive a data back, nothing recieved")
    }

    cancel()
    select {
    case <-done:
    case <-time.After(5 * time.Second):
        t.Error("expected to be done with the recorder, but it blocked")
    }

}
