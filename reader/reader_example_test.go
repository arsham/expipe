// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/lib"
)

func ExampleSimpleReader() {
	log := lib.DiscardLogger()
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"the key": "is the value!"}`)
	}))
	defer ts.Close()

	jobChan := make(chan context.Context, 10)
	errorChan := make(chan communication.ErrorMessage, 10)
	resultChan := make(chan *ReadJobResult, 10)
	red, _ := NewSimpleReader(log, ts.URL, jobChan, resultChan, errorChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
	stop := make(communication.StopChannel)
	red.Start(ctx, stop)

	// Issuing a job
	red.JobChan() <- communication.NewReadJob(ctx)

	// Lets check the errors
	select {
	case <-errorChan:
		panic("Wasn't expecting any errors")
	default:
		fmt.Println("No errors reported")
	}

	res := <-red.ResultChan()
	// Let's read what it retrieved
	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Res)
	fmt.Println("Result is:", buf.String())

	done := make(chan struct{})
	stop <- done
	<-done
	fmt.Println("Reader has finished")
	// We need to cancel the job now
	fmt.Println("All done!")
	// Output:
	// No errors reported
	// Result is: {"the key": "is the value!"}
	// Reader has finished
	// All done!
}

func ExampleSimpleReader_start() {
	log := lib.DiscardLogger()
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, `{"the key": "is the value!"}`) }))
	defer ts.Close()

	jobChan := make(chan context.Context)
	errorChan := make(chan communication.ErrorMessage)
	resultChan := make(chan *ReadJobResult)

	red, _ := NewSimpleReader(log, ts.URL, jobChan, resultChan, errorChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
	stop := make(communication.StopChannel)
	red.Start(ctx, stop)

	done := make(chan struct{})
	stop <- done
	<-done
	fmt.Println("Reader has stopped its event loop!")

	// Output:
	// Reader has stopped its event loop!
}
