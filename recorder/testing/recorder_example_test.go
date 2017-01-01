// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/recorder"
)

func ExampleSimpleRecorder() {
	var err error
	log := lib.DiscardLogger()
	ctx := context.Background()
	receivedPayload := make(chan string)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPayload <- "I have received the payload!"
	}))
	defer ts.Close()

	rec, _ := NewSimpleRecorder(ctx, log, "reader_example", ts.URL, "intexName", time.Second, 5)
	payload := datatype.NewContainer([]datatype.DataType{
		datatype.StringType{Key: "key", Value: "value"},
	})
	job := &recorder.RecordJob{
		Payload:   payload,
		IndexName: "my index",
		Time:      time.Now(),
	}
	// Issuing a job
	go func() {
		err = rec.Record(ctx, job)
		// Lets check the errors
		if err != nil {
			panic("Wasn't expecting any errors")
		}
	}()
	fmt.Println(<-receivedPayload)
	fmt.Println("No errors reported")

	// Issuing another job
	go rec.Record(ctx, job)
	fmt.Println(<-receivedPayload)

	// Output:
	// I have received the payload!
	// No errors reported
	// I have received the payload!
}
