// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expipe/engine"
	"github.com/arsham/expipe/reader"
	rdt "github.com/arsham/expipe/reader/testing"
	"github.com/arsham/expipe/recorder"
	rct "github.com/arsham/expipe/recorder/testing"
	"github.com/arsham/expipe/tools"
	"github.com/arsham/expipe/tools/token"
)

func BenchmarkEngineOnManyRecorders(b *testing.B) {
	bcs := []int{
		1,
		5,
		10,
		20,
		30,
	}
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {},
	))
	defer ts.Close()
	log := tools.DiscardLogger()
	log.Level = tools.ErrorLevel
	for _, bc := range bcs {
		name := fmt.Sprintf("Benchmark_%d", bc)
		b.Run(name, func(b *testing.B) {

			ctx, cancel := context.WithCancel(context.Background())

			// Setting the intervals to an hour so the benchmark can issue jobs
			red, err := rdt.New(
				reader.WithLogger(log),
				reader.WithEndpoint(ts.URL),
				reader.WithName("recorder_example"),
				reader.WithTypeName("typeName"),
				reader.WithTimeout(time.Hour),
			)
			job := token.New(ctx)
			jobID := job.ID()

			red.ReadFunc = func(*token.Context) (*reader.Result, error) {
				resp := &reader.Result{
					ID:       jobID,
					Content:  []byte(`{"lucifer":666}`),
					TypeName: red.TypeName(),
					Mapper:   red.Mapper(),
				}
				return resp, nil

			}

			if err != nil {
				b.Fatal(err)
			}
			recs, err := makeRecorders(bc, log, ts.URL)
			if err != nil {
				b.Fatal(err)
			}
			e, _ := engineWithReadRecs(ctx, log, red, recs)

			done := make(chan struct{})
			go func(done chan struct{}) {
				engine.Start(e)
				done <- struct{}{}
			}(done)

			for n := 0; n < b.N; n++ {
				if _, err := red.Read(token.New(ctx)); err != nil {
					b.Fatal(err)
				}
			}
			cancel()
			<-done
		})
	}
}

func makeRecorders(count int, log tools.FieldLogger, url string) (map[string]recorder.DataRecorder, error) {
	recs := make(map[string]recorder.DataRecorder, count)
	now := time.Now()
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("reader_%d", i)
		rec, err := rct.New(
			recorder.WithLogger(log),
			recorder.WithEndpoint(url),
			recorder.WithName(name),
			recorder.WithIndexName("example_index"),
			recorder.WithTimeout(time.Second),
		)
		if err != nil {
			return nil, err
		}
		rec.RecordFunc = func(c context.Context, job recorder.Job) error {
			p := new(bytes.Buffer)
			job.Payload.Generate(p, now)
			return nil
		}
		recs[rec.Name()] = rec
	}
	return recs, nil
}
