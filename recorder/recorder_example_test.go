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

	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/recorder"
	"github.com/arsham/expvastic/recorder/testing"
)

func ExampleDataRecorder() {
	ctx := context.Background()
	receivedPayload := make(chan string)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPayload <- "I have received the payload!"
	}))
	defer ts.Close()

	rec := testing.GetRecorder(ctx, ts.URL)
	payload := datatype.New([]datatype.DataType{
		datatype.StringType{Key: "key", Value: "value"},
	})
	job := &recorder.RecordJob{
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
	// I have received the payload!
	// No errors reported
	// I have received the payload!
}
