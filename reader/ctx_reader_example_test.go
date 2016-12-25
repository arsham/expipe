// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
)

func ExampleCtxReader_Get_a() {
	ctxReader := NewCtxReader("bad url")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res, err := ctxReader.Get(ctx)
	fmt.Println(res)
	fmt.Println(err != nil)
	// Output:
	// <nil>
	// true

}

func ExampleCtxReader_Get_b() {
	resp := "my response"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, resp)
	}))
	defer ts.Close()

	ctxReader := NewCtxReader(ts.URL)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	res, err := ctxReader.Get(ctx)
	defer res.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)

	fmt.Println("err == nil:", err == nil)
	fmt.Println("res != nil:", res != nil)
	fmt.Println("Response body:", buf.String())
	// Output:
	// err == nil: true
	// res != nil: true
	// Response body: my response
}
