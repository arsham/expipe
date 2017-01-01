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
	"github.com/arsham/expvastic/lib"
	reader_testing "github.com/arsham/expvastic/reader/testing"
	recorder_testing "github.com/arsham/expvastic/recorder/testing"
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

	red, err := reader_testing.NewSimpleReader(log, redTs.URL, "reader_example", "typeName", time.Millisecond, time.Millisecond, 5) //for testing
	if err != nil {
		panic(err)
	}
	rec, err := recorder_testing.NewSimpleRecorder(ctx, log, "reader_example", recTs.URL, "intexName", time.Millisecond, 5)
	if err != nil {
		panic(err)
	}

	e, err := expvastic.NewWithReadRecorder(ctx, log, rec, red)
	done := make(chan struct{})
	go func() {
		e.Start()
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
