// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expipe_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expipe"
	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/internal/token"
	"github.com/arsham/expipe/reader"
	reader_testing "github.com/arsham/expipe/reader/testing"
	recorder_testing "github.com/arsham/expipe/recorder/testing"
)

func BenchmarkEngineSingle(b *testing.B) {
	benchmarkEngineOnManyRecorders(1, b)
}

func BenchmarkEngineMulti2(b *testing.B) {
	benchmarkEngineOnManyRecorders(2, b)
}

func BenchmarkEngineMulti10(b *testing.B) {
	benchmarkEngineOnManyRecorders(10, b)
}

func BenchmarkEngineMulti20(b *testing.B) {
	benchmarkEngineOnManyRecorders(20, b)
}

func BenchmarkEngineMulti100(b *testing.B) {
	benchmarkEngineOnManyRecorders(100, b)
}

func benchmarkEngineOnManyRecorders(count int, b *testing.B) {
	bcs := []struct {
		readChanBuff, readResChanBuff, recChanBuff, recResChan, readers int
	}{
		{0, 0, 0, 0, 1},
		{0, 0, 0, 10, 10},
		{0, 0, 10, 0, 10},
		{0, 10, 0, 0, 10},
		{10, 0, 0, 0, 10},
		{0, 0, 10, 10, 10},
		{0, 10, 0, 10, 10},
		{10, 0, 0, 10, 10},
		{0, 10, 10, 10, 10},
		{10, 0, 10, 10, 100},
		{10, 10, 10, 10, 1000},
		{100, 100, 100, 100, 1000},
		{1000, 1000, 1000, 1000, 1000},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()
	log := internal.DiscardLogger()
	log.Level = internal.ErrorLevel
	for _, bc := range bcs {
		ctx, cancel := context.WithCancel(context.Background())
		name := fmt.Sprintf("Benchmark-%d_%d_%d_%d_(r:%d)", bc.readChanBuff, bc.readResChanBuff, bc.recChanBuff, bc.recResChan, bc.readers)

		// Setting the intervals to an hour so the benchmark can issue jobs
		rec, _ := recorder_testing.New(ctx, log, "reacorder_example", ts.URL, "intexName", time.Hour, 5)
		reds, err := makeReaders(ctx, bc.readers, log, ts.URL)
		if err != nil {
			b.Fatal(err)
		}
		e, _ := expipe.New(ctx, log, rec, reds)

		done := make(chan struct{})
		go func(done chan struct{}) {
			e.Start()
			done <- struct{}{}
		}(done)

		b.Run(name, func(b *testing.B) {
			benchmarkEngine(ctx, reds, b)
		})
		cancel()
		<-done
	}
}

func benchmarkEngine(ctx context.Context, reds map[string]reader.DataReader, b *testing.B) {
	for n := 0; n < b.N; n++ {
		for _, red := range reds {
			red.Read(token.New(ctx))
		}
	}
}

func makeReaders(ctx context.Context, count int, log internal.FieldLogger, url string) (map[string]reader.DataReader, error) {
	reds := make(map[string]reader.DataReader, count)
	readFunc := func(m *reader_testing.Reader) func(job *token.Context) (*reader.Result, error) {
		return func(job *token.Context) (*reader.Result, error) {
			res := &reader.Result{
				ID:       job.ID(),
				Time:     time.Now(),
				Content:  []byte(``),
				TypeName: m.TypeName(),
				Mapper:   m.Mapper(),
			}
			return res, nil
		}
	}
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("reader_%d", i)
		red, err := reader_testing.New(log, url, name, "example_type", time.Hour, time.Hour, 10)
		if err != nil {
			return nil, err
		}
		red.ReadFunc = readFunc(red)
		reds[red.Name()] = red
	}
	return reds, nil
}
