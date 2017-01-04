// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/arsham/expvastic/communication"
	reader "github.com/arsham/expvastic/reader/testing"
)

// This example shows the reader hits the endpoint when the Read method is called.
func ExampleDataReader_read() {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"the key": "is the value!"}`)
	}))
	defer ts.Close()

	red := reader.GetReader(ts.URL) // this is a mocked version, but the example's principals stays the same.
	err := red.Ping()
	fmt.Println("Ping errors:", err)

	job := communication.NewReadJob(context.Background())
	res, err := red.Read(job) // Issuing a job

	if err == nil { // Lets check the errors
		fmt.Println("No errors reported")
	}

	fmt.Println("Result is:", string(res.Res))

	// Output:
	// Ping errors: <nil>
	// No errors reported
	// Result is: {"the key": "is the value!"}
}
