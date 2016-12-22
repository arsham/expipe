// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/recorder"
)

// func BenchmarkEngineSingle(b *testing.B) {
// 	benchmarkEngineOnManyRecorders(1, b)
// }

// func BenchmarkEngineMulti2(b *testing.B) {
// 	benchmarkEngineOnManyRecorders(2, b)
// }

// func BenchmarkEngineMulti10(b *testing.B) {
// 	benchmarkEngineOnManyRecorders(10, b)
// }

// func BenchmarkEngineMulti20(b *testing.B) {
// 	benchmarkEngineOnManyRecorders(20, b)
// }

func BenchmarkEngineMulti100(b *testing.B) {
	benchmarkEngineOnManyRecorders(100, b)
}

func benchmarkEngineOnManyRecorders(count int, b *testing.B) {
	bcs := []struct {
		readChanBuff, readResChanBuff, recChanBuff, recResChan int
	}{
		// {0, 0, 0, 0},
		// {0, 0, 0, 10},
		// {0, 0, 10, 0},
		// {0, 10, 0, 0},
		// {10, 0, 0, 0},
		// {0, 0, 10, 10},
		// {0, 10, 0, 10},
		// {10, 0, 0, 10},
		// {0, 10, 10, 10},
		// {10, 0, 10, 10},
		{10, 10, 10, 10},
		{100, 100, 100, 100},
		{1000, 1000, 1000, 1000},
	}
	for _, bc := range bcs {
		name := fmt.Sprintf("Benchmak-%d_%d_%d_%d", bc.readChanBuff, bc.readResChanBuff, bc.recChanBuff, bc.recResChan)
		log := lib.DiscardLogger()
		jobChan := make(chan context.Context, bc.readChanBuff)
		resultChan := make(chan *reader.ReadJobResult, bc.readResChanBuff)

		ctx, cancel := context.WithCancel(context.Background())
		redTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, `{"the key": "is the value!"}`) }))
		recTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer redTs.Close()
		defer recTs.Close()

		ctxReader := reader.NewCtxReader(redTs.URL)
		// Settig the intervals to an hour so the benchmark can issue jobs
		red, _ := reader.NewSimpleReader(log, ctxReader, jobChan, resultChan, "reader_example", "example_type", time.Hour, time.Hour)
		red.Start(ctx)
		recs := makeRecorders(ctx, 1, log, bc.recChanBuff, recTs.URL)
		cl, _ := expvastic.NewWithReadRecorder(ctx, log, bc.recResChan, red, recs...)
		done := cl.Start()
		b.Run(name, func(b *testing.B) {
			benchmarkEngine(ctx, red, b)
		})
		cancel()
		<-done
	}
}

func benchmarkEngine(ctx context.Context, red *reader.SimpleReader, b *testing.B) {
	for n := 0; n < b.N; n++ {
		red.JobChan() <- ctx
	}
}

func makeRecorders(ctx context.Context, count int, log logrus.FieldLogger, chanBuff int, url string) []recorder.DataRecorder {
	recs := make([]recorder.DataRecorder, count)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("recorder_%d", i)
		payloadChan := make(chan *recorder.RecordJob, chanBuff)
		rec, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, name, url, "intexName", time.Hour, time.Hour)
		rec.Start(ctx)
		recs[i] = rec
	}
	return recs
}
