// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/lib"
)

func ExampleSimpleRecorder() {
	log := lib.DiscardLogger()
	ctx := context.Background()
	receivedPayload := make(chan string)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPayload <- "I have received the payload!"
	}))
	defer ts.Close()

	payloadChan := make(chan *RecordJob)
	errorChan := make(chan communication.ErrorMessage)
	rec, _ := NewSimpleRecorder(ctx, log, payloadChan, errorChan, "reader_example", ts.URL, "intexName", time.Second)
	stop := make(communication.StopChannel)
	rec.Start(ctx, stop)
	payload := datatype.NewContainer([]datatype.DataType{
		datatype.StringType{Key: "key", Value: "value"},
	})

	job := &RecordJob{
		Ctx:       ctx,
		Payload:   payload,
		IndexName: "my index",
		Time:      time.Now(),
	}
	// Issuing a job
	rec.PayloadChan() <- job
	fmt.Println(<-receivedPayload)
	// Lets check the errors
	select {
	case <-errorChan:
		panic("Wasn't expecting any errors")
	default:
		fmt.Println("No errors reported")
	}

	// Issuing another job
	rec.PayloadChan() <- job
	fmt.Println(<-receivedPayload)

	// The recorder should finish gracefully
	done := make(chan struct{})
	stop <- done
	<-done
	fmt.Println("Reader has finished")

	// Output:
	// I have received the payload!
	// No errors reported
	// I have received the payload!
	// Reader has finished
}

func ExampleSimpleRecorder_start() {
	log := lib.DiscardLogger()
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"the key": "is the value!"}`)
	}))
	defer ts.Close()

	payloadChan := make(chan *RecordJob)
	errorChan := make(chan communication.ErrorMessage)
	rec, _ := NewSimpleRecorder(ctx, log, payloadChan, errorChan, "reader_example", ts.URL, "intexName", 10*time.Millisecond)
	stop := make(communication.StopChannel)
	rec.Start(ctx, stop)

	done := make(chan struct{})
	stop <- done
	<-done
	fmt.Println("Recorder has stopped its event loop!")
	// Output:
	// Recorder has stopped its event loop!
}
