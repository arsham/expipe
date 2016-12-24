// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/lib"
)

func BenchmarkRecorder0_0(b *testing.B)       { benchmarkRecorder(0, 0, b) }
func BenchmarkRecorder0_10(b *testing.B)      { benchmarkRecorder(0, 10, b) }
func BenchmarkRecorder10_0(b *testing.B)      { benchmarkRecorder(10, 0, b) }
func BenchmarkRecorder20_20(b *testing.B)     { benchmarkRecorder(20, 20, b) }
func BenchmarkRecorder100_100(b *testing.B)   { benchmarkRecorder(100, 100, b) }
func BenchmarkRecorder100_10(b *testing.B)    { benchmarkRecorder(100, 10, b) }
func BenchmarkRecorder10_100(b *testing.B)    { benchmarkRecorder(10, 100, b) }
func BenchmarkRecorder1000_1000(b *testing.B) { benchmarkRecorder(1000, 1000, b) }
func BenchmarkRecorder1000_0(b *testing.B)    { benchmarkRecorder(1000, 0, b) }
func BenchmarkRecorder0_1000(b *testing.B)    { benchmarkRecorder(0, 1000, b) }

func benchmarkRecorder(jobBuffC, doneBuffC int, b *testing.B) {
	log := lib.DiscardLogger()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	payloadChan := make(chan *RecordJob, jobBuffC)
	errorChan := make(chan communication.ErrorMessage, doneBuffC)
	rec, _ := NewSimpleRecorder(ctx, log, payloadChan, errorChan, "reader_example", ts.URL, "intexName", 10*time.Millisecond)
	rec.Start(ctx)

	for n := 0; n < b.N; n++ {
		job := &RecordJob{
			Ctx:       ctx,
			Payload:   nil,
			IndexName: "my index",
			Time:      time.Now(),
		}
		rec.PayloadChan() <- job
		<-errorChan
	}
}
