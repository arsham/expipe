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
	t.Parallel()
	log := lib.DiscardLogger()
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"the key": "is the value!"}`)
	}))
	defer ts.Close()

	jobChan := make(chan context.Context)
	errorChan := make(chan communication.ErrorMessage)
	resultChan := make(chan *ReadJobResult, 1)
	red, _ := NewSimpleReader(log, NewCtxReader(ts.URL), jobChan, resultChan, errorChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
	stop := make(communication.StopChannel)
	red.Start(ctx, stop)

	select {
	case red.JobChan() <- communication.NewReadJob(ctx):
	case <-time.After(5 * time.Second):
		t.Error("expected the reader to receive the job, but it blocked")
	}
	done := make(chan struct{})
	stop <- done
	<-done

}

func TestSimpleReaderSendsResult(t *testing.T) {
	t.Parallel()
	var res *ReadJobResult
	log := lib.DiscardLogger()
	ctx := context.Background()

	desire := `{"the key": "is the value!"}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, desire)
	}))
	defer ts.Close()

	jobChan := make(chan context.Context)
	errorChan := make(chan communication.ErrorMessage)
	resultChan := make(chan *ReadJobResult)
	red, _ := NewSimpleReader(log, NewCtxReader(ts.URL), jobChan, resultChan, errorChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
	stop := make(communication.StopChannel)
	red.Start(ctx, stop)

	red.JobChan() <- communication.NewReadJob(ctx)

	select {
	case err := <-errorChan:
		t.Errorf("didn't expect errors, got (%v)", err.Error())
	case <-time.After(20 * time.Millisecond):
	}

	select {
	case res = <-resultChan:
	case <-time.After(5 * time.Second):
		t.Error("expected to receive a data back, nothing received")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Res)
	if buf.String() != desire {
		t.Errorf("want (%s), got (%s)", desire, buf.String())
	}
	done := make(chan struct{})
	stop <- done
	<-done

}

func TestSimpleReaderReadsOnBufferedChan(t *testing.T) {
	t.Parallel()
	var res *ReadJobResult
	log := lib.DiscardLogger()
	ctx := context.Background()
	desire := `{"the key": "is the value!"}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, desire)
	}))
	defer ts.Close()

	jobChan := make(chan context.Context, 10)
	errorChan := make(chan communication.ErrorMessage, 10)
	resultChan := make(chan *ReadJobResult)

	red, _ := NewSimpleReader(log, NewCtxReader(ts.URL), jobChan, resultChan, errorChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
	stop := make(communication.StopChannel)
	red.Start(ctx, stop)

	red.JobChan() <- communication.NewReadJob(ctx)

	select {
	case err := <-errorChan:
		t.Errorf("didn't expect errors, got (%v)", err.Error())
	case <-time.After(20 * time.Millisecond):
	}

	select {
	case res = <-resultChan:
	case <-time.After(5 * time.Second):
		t.Error("expected to receive a data back, nothing received")
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
	done := make(chan struct{})
	stop <- done
	<-done

}

func TestSimpleReaderDrainsAfterClosingContext(t *testing.T) {
	t.Parallel()
	var res *ReadJobResult
	log := lib.DiscardLogger()
	ctx := context.Background()
	desire := `{"the key": "is the value!"}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, desire)
	}))
	defer ts.Close()

	jobChan := make(chan context.Context, 10)
	errorChan := make(chan communication.ErrorMessage, 10)
	resultChan := make(chan *ReadJobResult)

	red, _ := NewSimpleReader(log, NewCtxReader(ts.URL), jobChan, resultChan, errorChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
	stop := make(communication.StopChannel)
	red.Start(ctx, stop)

	red.JobChan() <- communication.NewReadJob(ctx)

	select {
	case err := <-errorChan:
		t.Errorf("didn't expect errors, got (%v)", err.Error())
	case <-time.After(20 * time.Millisecond):
	}

	select {
	case res = <-resultChan:
	case <-time.After(5 * time.Second):
		t.Error("expected to receive a data back, nothing received")
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
	done := make(chan struct{})
	stop <- done
	<-done
}

func TestSimpleReaderCloses(t *testing.T) {
	t.Parallel()
	var res *ReadJobResult
	log := lib.DiscardLogger()
	ctx := context.Background()
	desire := `{"the key": "is the value!"}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, desire)
	}))
	defer ts.Close()

	jobChan := make(chan context.Context)
	errorChan := make(chan communication.ErrorMessage)
	resultChan := make(chan *ReadJobResult)
	red, _ := NewSimpleReader(log, NewCtxReader(ts.URL), jobChan, resultChan, errorChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
	stop := make(communication.StopChannel)
	red.Start(ctx, stop)

	red.JobChan() <- communication.NewReadJob(ctx)
	res = <-resultChan
	defer res.Res.Close()
	done := make(chan struct{})
	stop <- done

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected to be done with the reader, but it blocked")
	}
}

func TestSimpleReaderClosesWithBufferedChans(t *testing.T) {
	t.Parallel()
	var res *ReadJobResult
	log := lib.DiscardLogger()
	ctx := context.Background()
	desire := `{"the key": "is the value!"}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, desire)
	}))
	defer ts.Close()

	jobChan := make(chan context.Context, 1000)
	errorChan := make(chan communication.ErrorMessage, 1000)
	resultChan := make(chan *ReadJobResult, 1000)
	red, _ := NewSimpleReader(log, NewCtxReader(ts.URL), jobChan, resultChan, errorChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
	stop := make(communication.StopChannel)
	red.Start(ctx, stop)

	red.JobChan() <- communication.NewReadJob(ctx)
	res = <-resultChan
	defer res.Res.Close()

	done := make(chan struct{})
	stop <- done
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected to be done with the reader, but it blocked")
	}
}
