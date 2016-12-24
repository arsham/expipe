// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/lib"
)

var count = 0

func BenchmarkReader0_0(b *testing.B)           { benchmarkReader(0, 0, b) }
func BenchmarkReader0_10(b *testing.B)          { benchmarkReader(0, 10, b) }
func BenchmarkReader10_0(b *testing.B)          { benchmarkReader(10, 0, b) }
func BenchmarkReader20_20(b *testing.B)         { benchmarkReader(20, 20, b) }
func BenchmarkReader100_100(b *testing.B)       { benchmarkReader(100, 100, b) }
func BenchmarkReader100_10(b *testing.B)        { benchmarkReader(100, 10, b) }
func BenchmarkReader10_100(b *testing.B)        { benchmarkReader(10, 100, b) }
func BenchmarkReader1000_1000(b *testing.B)     { benchmarkReader(1000, 1000, b) }
func BenchmarkReader1000_0(b *testing.B)        { benchmarkReader(1000, 0, b) }
func BenchmarkReader0_1000(b *testing.B)        { benchmarkReader(0, 1000, b) }
func BenchmarkReader100000_100000(b *testing.B) { benchmarkReader(100000, 100000, b) }

func benchmarkReader(jobBuffC, resBuffC int, b *testing.B) {
	log := lib.DiscardLogger()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"the key": "is the value!"}`)
	}))
	defer ts.Close()

	jobChan := make(chan context.Context, jobBuffC)
	errorChan := make(chan communication.ErrorMessage, jobBuffC)
	resultChan := make(chan *ReadJobResult, resBuffC)
	ctxReader := NewCtxReader(ts.URL)
	red, _ := NewSimpleReader(log, ctxReader, jobChan, resultChan, errorChan, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond)
	red.Start(ctx)

	for n := 0; n < b.N; n++ {
		red.JobChan() <- communication.NewReadJob(ctx)
		<-red.ResultChan()
	}
}
