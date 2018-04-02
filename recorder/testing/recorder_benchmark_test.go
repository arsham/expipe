// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expipe/recorder"
	"github.com/arsham/expipe/tools"
)

func BenchmarkRecorder(b *testing.B) {
	log := tools.DiscardLogger()
	ctx := context.Background()

	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)
	defer ts.Close()

	rec, err := New(
		recorder.WithLogger(log),
		recorder.WithEndpoint(ts.URL),
		recorder.WithName("recorder_example"),
		recorder.WithIndexName("recorder_example"),
		recorder.WithTimeout(time.Second),
		recorder.WithBackoff(10),
	)
	if err != nil {
		b.Fatal(err)
	}
	err = rec.Ping()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		job := &recorder.Job{
			Payload:   nil,
			IndexName: "my index",
			Time:      time.Now(),
		}
		rec.Record(ctx, job)
	}
}
