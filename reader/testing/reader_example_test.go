// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/lib"
)

func ExampleSimpleReader_read() {
	log := lib.DiscardLogger()
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"the key": "is the value!"}`)
	}))
	defer ts.Close()

	red, _ := NewSimpleReader(log, ts.URL, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond, 10)

	// Issuing a job
	res, err := red.Read(communication.NewReadJob(ctx))
	// Lets check the errors
	if err == nil {
		fmt.Println("No errors reported")
	}

	// Let's read what it retrieved
	fmt.Println("Result is:", string(res.Res))

	// Output:
	// No errors reported
	// Result is: {"the key": "is the value!"}
}
