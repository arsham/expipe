// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expipe/engine"
	"github.com/arsham/expipe/reader"
	rtd "github.com/arsham/expipe/reader/testing"
	"github.com/arsham/expipe/recorder"
	rct "github.com/arsham/expipe/recorder/testing"
	"github.com/arsham/expipe/tools"
	"github.com/arsham/expipe/tools/token"
)

// TODO: benchmark on many readers.

func BenchmarkEngineOnManyRecorders(b *testing.B) {
	bcs := []int{
		1,
		10,
		100,
		1000,
		10000,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()
	log := tools.DiscardLogger()
	log.Level = tools.ErrorLevel
	for _, bc := range bcs {
		name := fmt.Sprintf("Benchmark-%d", bc)
		b.Run(name, func(b *testing.B) {

			ctx, cancel := context.WithCancel(context.Background())

			// Setting the intervals to an hour so the benchmark can issue jobs
			rec, _ := rct.New(
				recorder.WithLogger(log),
				recorder.WithEndpoint(ts.URL),
				recorder.WithName("recorder_example"),
				recorder.WithIndexName("indexName"),
				recorder.WithTimeout(time.Hour),
				recorder.WithBackoff(5),
			)
			reds, err := makeReaders(bc, log, ts.URL)
			if err != nil {
				b.Fatal(err)
			}
			e, _ := engineWithReadRecs(ctx, log, rec, reds)

			done := make(chan struct{})
			go func(done chan struct{}) {
				engine.Start(e)
				done <- struct{}{}
			}(done)

			benchmarkEngine(ctx, reds, b)
			cancel()
			<-done
		})
	}
}

// TODO: refactor this benchmark.
func benchmarkEngine(ctx context.Context, reds map[string]reader.DataReader, b *testing.B) {
	for n := 0; n < b.N; n++ {
		for _, red := range reds {
			if _, err := red.Read(token.New(ctx)); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func makeReaders(count int, log tools.FieldLogger, url string) (map[string]reader.DataReader, error) {
	reds := make(map[string]reader.DataReader, count)
	readFunc := func(m *rtd.Reader) func(job *token.Context) (*reader.Result, error) {
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
		red, err := rtd.New(
			reader.WithLogger(log),
			reader.WithEndpoint(url),
			reader.WithName(name),
			reader.WithTypeName("example_type"),
			reader.WithInterval(time.Hour),
			reader.WithTimeout(time.Second),
			reader.WithBackoff(10),
		)
		if err != nil {
			return nil, err
		}
		red.ReadFunc = readFunc(red)
		reds[red.Name()] = red
	}
	return reds, nil
}
