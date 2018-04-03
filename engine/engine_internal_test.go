// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expipe/reader"
	rdt "github.com/arsham/expipe/reader/testing"
	"github.com/arsham/expipe/recorder"
	rct "github.com/arsham/expipe/recorder/testing"
	"github.com/arsham/expipe/tools"
	"github.com/arsham/expipe/tools/token"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus/hooks/test"
)

var (
	log        tools.FieldLogger
	testServer *httptest.Server
)

func init() {
	log = tools.DiscardLogger()
	testServer = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)
}

func withRecorder(ctx context.Context, log tools.FieldLogger) (*Engine, error) {
	rec, _ := rct.New(
		recorder.WithLogger(log),
		recorder.WithEndpoint(testServer.URL),
		recorder.WithName("recorder_test"),
		recorder.WithIndexName("indexName"),
		recorder.WithTimeout(time.Hour),
		recorder.WithBackoff(5),
	)

	err := rec.Ping()
	if err != nil {
		return nil, err
	}
	return &Engine{
		name:       "test_engine",
		ctx:        ctx,
		log:        log,
		recorder:   rec,
		readerJobs: make(chan *reader.Result),
	}, nil
}

// setReaders is used only in tests.
func (e *Engine) setReaders(readers map[string]reader.DataReader) {
	e.redmu.Lock()
	defer e.redmu.Unlock()
	e.readers = readers
}

func TestRemoveReader(t *testing.T) {
	endpoint := "http://localhost"
	r1, err := rdt.New(reader.WithName("r1"), reader.WithEndpoint(endpoint))
	if err != nil {
		t.Fatalf("rdt.New(name: r1): err = (%#v); want (nil)", err)
	}
	r2, err := rdt.New(reader.WithName("r2"), reader.WithEndpoint(endpoint))
	if err != nil {
		t.Fatalf("rdt.New(name: r2): err = (%#v); want (nil)", err)
	}
	r3, err := rdt.New(reader.WithName("r3"), reader.WithEndpoint(endpoint))
	if err != nil {
		t.Fatalf("rdt.New(name: r3): err = (%#v); want (nil)", err)
	}
	r4, err := rdt.New(reader.WithName("r4"), reader.WithEndpoint(endpoint))
	if err != nil {
		t.Fatalf("rdt.New(name: r4): err = (%#v); want (nil)", err)
	}
	e := &Engine{}
	e.setReaders(map[string]reader.DataReader{
		r1.Name(): r1,
		r2.Name(): r2,
		r3.Name(): r3,
		r4.Name(): r4,
	})
	e.removeReader(r1)
	if _, ok := e.readers[r1.Name()]; ok {
		t.Errorf("e.readers[r1.Name()]: didn't expect to have (%v) in the map", r1)
	}
	for _, r := range []reader.DataReader{r2, r3, r4} {
		if _, ok := e.readers[r.Name()]; !ok {
			t.Errorf("e.readers[r.Name()]: didn't expect it to remove (%v)", r)
		}
	}
}

func TestEventLoopCatchesReaderError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestEventLoopCatchesReaderError count in short mode")
	}
	t.Parallel()
	log, _ := test.NewNullLogger()

	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red, err := rdt.New(
		reader.WithLogger(tools.DiscardLogger()),
		reader.WithEndpoint(testServer.URL),
		reader.WithName("reader_name"),
		reader.WithTypeName("typeName"),
		reader.WithInterval(10*time.Millisecond),
		reader.WithTimeout(time.Second),
		reader.WithBackoff(5),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil): unexpected error occurred during reader creation", err)
	}
	if err = red.Ping(); err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}

	e.setReaders(map[string]reader.DataReader{red.Name(): red})

	errMsg := errMessage("an error happened")
	recorded := make(chan struct{})

	// Testing the engine catches errors
	red.ReadFunc = func(job *token.Context) (*reader.Result, error) {
		recorded <- struct{}{}
		return nil, errMsg
	}

	errChan := make(chan JobError)
	stop := make(chan struct{})
	e.issueReaderJob(red, errChan, stop)

	select {
	case <-recorded:
	case <-time.After(5 * time.Second):
		t.Error("expected to record, didn't happen")
	}
	cancel()
	select {
	case err := <-errChan:
		if errors.Cause(err.Err) != errMsg {
			t.Errorf("want (%v), got (%v)", errMsg, err.Err)
		}
	case <-time.After(5 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}
}
