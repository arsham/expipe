// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/arsham/expipe/engine"
	"github.com/arsham/expipe/tools"
)

func ExampleEngine_sendingJobs() {
	log := tools.DiscardLogger()
	ctx, cancel := context.WithCancel(context.Background())
	recorded := make(chan string)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorded <- "Job was recorded"
	}))
	defer ts.Close()

	red, redTearDown := getReader(log)
	defer redTearDown()
	rec := getRecorder(log, ts.URL)
	e, err := engineWithReadRecs(ctx, log, rec, red)
	if err != nil {
		panic(err)
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
