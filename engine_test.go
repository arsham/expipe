// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/recorder"
)

func TestEngineSendJob(t *testing.T) {
	var res *reader.ReadJobResult
	log := lib.DiscardLogger()
	ctx, cancel := context.WithCancel(context.Background())
	desire := `{"the key": "is the value!"}`
	recorded := false

	redTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, desire)
	}))
	defer redTs.Close()

	recTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorded = true
	}))
	defer recTs.Close()

	ctxReader := reader.NewCtxReader(redTs.URL)
	red, _ := reader.NewSimpleReader(log, ctxReader, "reader_example", 1*time.Millisecond, 1*time.Millisecond)
	rec, _ := recorder.NewSimpleRecorder(ctx, log, "reader_example", recTs.URL, "intexName", "typeName", 1*time.Millisecond, 1*time.Millisecond)
	redDone := red.Start()
	recDone := rec.Start()

	cl, err := expvastic.NewWithReadRecorder(ctx, log, red, rec)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	clDone := cl.Start()

	select {
	case red.JobChan() <- ctx:
	case <-time.After(time.Second):
		t.Error("expected the reader to recive the job, but it blocked")
	}

	select {
	case res = <-red.ResultChan():
		if res.Err != nil {
			t.Fatalf("want (nil), got (%v)", res.Err)
		}
	case <-time.After(5 * time.Second): // Should be more than the interval, otherwise the response is not ready yet
		//TODO: investigate
		t.Error("expected to recive a data back, nothing recieved")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Res)
	if buf.String() != desire {
		t.Errorf("want (%s), got (%s)", desire, buf.String())
	}

	if !recorded {
		t.Errorf("recorder didn't record the request")
	}

	cancel()
	cl.Stop()
	close(red.JobChan())
	close(rec.PayloadChan())
	if _, ok := <-redDone; ok {
		t.Error("expected the channel to be closed")
	}
	if _, ok := <-recDone; ok {
		t.Error("expected the channel to be closed")
	}
	if v, ok := <-clDone; ok {
		t.Error("expected the channel to be closed", v)
	}
}
