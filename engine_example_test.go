// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/recorder"
)

func ExampleEngine_sendingJobs() {
	log := lib.DiscardLogger()
	ctx, cancel := context.WithCancel(context.Background())
	recorded := make(chan string)

	redTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		desire := `{"the key": "is the value!"}`
		io.WriteString(w, desire)
	}))
	defer redTs.Close()

	recTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorded <- "Job was recorded"
	}))
	defer recTs.Close()

	jobChan := make(chan context.Context)
	errorChan := make(chan communication.ErrorMessage, 10)
	resultChan := make(chan *reader.ReadJobResult)
	payloadChan := make(chan *recorder.RecordJob)

	ctxReader := reader.NewCtxReader(redTs.URL)
	red, _ := reader.NewSimpleReader(log, ctxReader, jobChan, resultChan, errorChan, "reader_example", "typeName", time.Hour, time.Hour) // We want to issue manually
	rec, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, errorChan, "reader_example", recTs.URL, "intexName", time.Hour)

	e, err := expvastic.NewWithReadRecorder(ctx, log, errorChan, resultChan, rec, red)
	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()
	fmt.Println("Engine creation success:", err == nil)

	select {
	case jobChan <- communication.NewReadJob(ctx):
		fmt.Println("Just sent a job request")
	case <-time.After(1 * time.Second):
		panic("expected the reader to receive the job, but it blocked")
	}

	fmt.Println(<-recorded)

	select {
	case <-errorChan:
		panic("expected no errors")
	case <-time.After(10 * time.Millisecond):
		fmt.Println("No errors reported!")
	}
	// We can check again
	// Both readers and recorders produce errors if they need to
	select {
	case <-errorChan:
		panic("expected no errors")
	case <-time.After(10 * time.Millisecond):
		fmt.Println("No errors reported!")
	}

	cancel()
	<-done
	fmt.Println("Client closed gracefully")

	// Output:
	// Engine creation success: true
	// Just sent a job request
	// Job was recorded
	// No errors reported!
	// No errors reported!
	// Client closed gracefully
}
