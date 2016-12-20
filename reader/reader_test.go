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

    "github.com/arsham/expvastic/lib"
)

func TestSimpleReader(t *testing.T) {
    log := lib.DiscardLogger()
    ctx, _ := context.WithCancel(context.Background())
    desire := `{"the key": "is the value!"}`
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        io.WriteString(w, desire)
    }))
    defer ts.Close()

    ctxReader := NewCtxReader(ts.URL)
    rdr, _ := NewSimpleReader(log, ctxReader, "reader_example")
    done := rdr.Start()

    job, _ := context.WithCancel(ctx)
    select {
    case rdr.JobChan() <- job:
    case <-time.After(time.Second):
        t.Error("expected the reader to recive the job, but it blocked")
    }

    var res *ReadJobResult

    select {
    case res = <-rdr.ResultChan():
        if res.Err != nil {
            t.Errorf("want (nil), got (%v)", res.Err)
        }
    case <-time.After(time.Second):
        t.Error("expected to recive a data back, nothing recieved")
    }

    buf := new(bytes.Buffer)
    buf.ReadFrom(res.Res)
    if buf.String() != desire {
        t.Errorf("want (%s), got (%s)", desire, buf.String())
    }

    close(rdr.JobChan())
    select {
    case <-done:
    case <-time.After(time.Second):
        t.Error("expected to be done with the reader, but it blocked")
    }

}
