// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/tools"
	"github.com/arsham/expipe/tools/token"
)

func BenchmarkReader(b *testing.B) {
	log := tools.DiscardLogger()
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"the key": "is the value!"}`)
	}))
	defer ts.Close()

	red, err := New(
		reader.WithLogger(log),
		reader.WithEndpoint(ts.URL),
		reader.WithName("reader_example"),
		reader.WithTypeName("reader_example"),
		reader.WithInterval(10*time.Millisecond),
		reader.WithTimeout(time.Second),
		reader.WithBackoff(10),
	)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		red.Read(token.New(ctx))
	}
}
