// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/recorder"
	"github.com/arsham/expipe/recorder/testing"
)

// This example shows when a record job is issued, the recorder hits the endpoint.
func ExampleDataRecorder() {
	ctx := context.Background()
	receivedPayload := make(chan string)
	pinged := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !pinged {
			pinged = true
			return
		}
		receivedPayload <- "I have received the payload!"
	}))
	defer ts.Close()

	rec := testing.GetRecorder(ts.URL)
	rec.Ping()
	fmt.Println("Pinging successful")
	payload := datatype.New([]datatype.DataType{
		datatype.StringType{Key: "key", Value: "value"},
	})
	job := &recorder.Job{
		Payload:   payload,
		IndexName: "my index",
		Time:      time.Now(),
	}

	go func() {
		err := rec.Record(ctx, job)
		if err != nil {
			panic("Wasn't expecting any errors")
		}
	}()
	fmt.Println(<-receivedPayload)
	fmt.Println("No errors reported")

	go rec.Record(ctx, job) // Issuing another job
	fmt.Println(<-receivedPayload)

	// Output:
	// Pinging successful
	// I have received the payload!
	// No errors reported
	// I have received the payload!
}
