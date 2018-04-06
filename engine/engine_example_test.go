// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"time"

	"github.com/arsham/expipe/engine"
	"github.com/arsham/expipe/reader"
	rdt "github.com/arsham/expipe/reader/testing"
	"github.com/arsham/expipe/recorder"
	rct "github.com/arsham/expipe/recorder/testing"
	"github.com/arsham/expipe/tools"
)

func recorderWithURL(url string) recorder.DataRecorder {
	log := tools.DiscardLogger()
	rec, err := rct.New(
		recorder.WithLogger(log),
		recorder.WithEndpoint(url),
		recorder.WithName("recorder_example"),
		recorder.WithIndexName("indexName"),
		recorder.WithTimeout(time.Second),
	)
	if err != nil {
		log.Fatalln("This error should not happen:", err)
	}
	return rec
}

func readerWithURL(url string) reader.DataReader {
	log := tools.DiscardLogger()
	red, err := rdt.New(
		reader.WithLogger(log),
		reader.WithEndpoint(url),
		reader.WithName("reader_example"),
		reader.WithTypeName("typeName"),
		reader.WithInterval(time.Millisecond*100),
		reader.WithTimeout(time.Second),
	)
	if err != nil {
		log.Fatalln("This error should not happen:", err)
	}
	return red
}

// You need at least a pair of DataReader and DataRecorder to start an engine.
// In this example we are using the mocked versions.
func ExampleStart() {
	log := tools.DiscardLogger()
	ctx, cancel := context.WithCancel(context.Background())
	recorded := make(chan string)

	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			recorded <- "Job was recorded"
		},
	))
	defer ts.Close()

	red := getReader(log)
	rec := getRecorders(log, ts.URL)
	e, err := engineWithReadRecs(ctx, log, red, rec)
	if err != nil {
		log.Fatalln("This error should not happen:", err)
	}
	done := make(chan struct{})
	go func() {
		engine.Start(e)
		done <- struct{}{}
	}()
	fmt.Println("Engine creation success:", err == nil)
	fmt.Println(<-recorded)

	cancel()
	<-done
	fmt.Println("Client closed gracefully")

	// Output:
	// Engine creation success: true
	// Job was recorded
	// Client closed gracefully
}

// You can pass your configuration.
func ExampleNew() {
	log := tools.DiscardLogger()
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {},
	))
	ctx := context.Background()

	rec := recorderWithURL(ts.URL)
	red := readerWithURL(ts.URL)

	e, err := engine.New(
		engine.WithCtx(ctx),
		engine.WithLogger(log),
		engine.WithReader(red),
		engine.WithRecorders(rec),
	)
	fmt.Println("Error:", err)
	fmt.Println("Engine is nil:", e == nil)

	// Output:
	// Error: <nil>
	// Engine is nil: false
}

// Please note that if you have a duplicate, the last one will replace the
// old ones.
func ExampleNew_replaces() {
	log := tools.DiscardLogger()
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {},
	))
	ctx1, cancel := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel()
	defer cancel2()

	rec := recorderWithURL(ts.URL)
	red := readerWithURL(ts.URL)

	e, err := engine.New(
		engine.WithCtx(ctx1),
		engine.WithCtx(ctx2),
		engine.WithLogger(log),
		engine.WithReader(red),
		engine.WithRecorders(rec),
	)
	fmt.Println("Error:", err)
	fmt.Println("e.Ctx() == ctx1:", e.Ctx() == ctx1)
	fmt.Println("e.Ctx() == ctx2:", e.Ctx() == ctx2)

	// Output:
	// Error: <nil>
	// e.Ctx() == ctx1: false
	// e.Ctx() == ctx2: true
}

func ExampleWithCtx() {
	ctx := context.Background()
	o := &engine.Operator{}
	err := engine.WithCtx(ctx)(o)
	fmt.Println("Error:", err)
	fmt.Println("o.Ctx() == ctx:", o.Ctx() == ctx)

	// Output:
	// Error: <nil>
	// o.Ctx() == ctx: true
}

func ExampleWithLogger() {
	log := tools.DiscardLogger()
	o := &engine.Operator{}
	err := engine.WithLogger(log)(o)
	fmt.Println("Error:", err)
	fmt.Println("o.Log() == log:", o.Log() == log)

	// Output:
	// Error: <nil>
	// o.Log() == log: true
}

func ExampleWithRecorders() {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {},
	))
	rec := recorderWithURL(ts.URL)
	o := &engine.Operator{}
	err := engine.WithRecorders(rec)(o)
	fmt.Println("Error:", err)

	// Output:
	// Error: <nil>
}

// If the DataRecorder couldn't ping, it will return an error.
func ExampleWithRecorders_pingError() {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {},
	))
	ts.Close()
	rec := recorderWithURL(ts.URL)
	o := &engine.Operator{}
	err := engine.WithRecorders(rec)(o)
	fmt.Println("Error type:", reflect.TypeOf(err))

	// Output:
	// Error type: engine.PingError
}

func ExampleWithReader() {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {},
	))
	red := readerWithURL(ts.URL)

	o := &engine.Operator{}
	err := engine.WithReader(red)(o)
	fmt.Println("Error:", err)

	// Output:
	// Error: <nil>
}

// If the DataReader couldn't ping, it will return an error.
func ExampleWithReader_pingError() {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {},
	))
	ts.Close()
	red := readerWithURL(ts.URL)

	o := &engine.Operator{}
	err := engine.WithReader(red)(o)
	fmt.Println("Error type:", reflect.TypeOf(err))

	// Output:
	// Error type: engine.PingError
}
