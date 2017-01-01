// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	reader_testing "github.com/arsham/expvastic/reader/testing"
	recorder_testing "github.com/arsham/expvastic/recorder/testing"
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
	log := lib.DiscardLogger()
	log.Level = logrus.ErrorLevel
	for _, bc := range bcs {
		ctx, cancel := context.WithCancel(context.Background())
		name := fmt.Sprintf("Benchmark-%d_%d_%d_%d_(r:%d)", bc.readChanBuff, bc.readResChanBuff, bc.recChanBuff, bc.recResChan, bc.readers)

		// Setting the intervals to an hour so the benchmark can issue jobs
		rec, _ := recorder_testing.NewSimpleRecorder(ctx, log, "reacorder_example", "http://127.0.0.1", "intexName", time.Hour, 5)
		reds := makeReaders(ctx, bc.readers, log, "http://127.0.0.1")
		e, _ := expvastic.NewWithReadRecorder(ctx, log, rec, reds...)

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

func benchmarkEngine(ctx context.Context, reds []reader.DataReader, b *testing.B) {
	for n := 0; n < b.N; n++ {
		for _, red := range reds {
			red.Read(communication.NewReadJob(ctx))
		}
	}
}

func makeReaders(ctx context.Context, count int, log logrus.FieldLogger, url string) []reader.DataReader {
	reds := make([]reader.DataReader, count)
	readFunc := func(m *reader_testing.SimpleReader) func(ctx context.Context) (*reader.ReadJobResult, error) {
		return func(job context.Context) (*reader.ReadJobResult, error) {
			id := communication.JobValue(job)
			res := &reader.ReadJobResult{
				ID:       id,
				Time:     time.Now(),
				Res:      []byte(``),
				TypeName: m.TypeName(),
				Mapper:   m.Mapper(),
			}
			return res, nil
		}
	}
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("reader_%d", i)
		red, _ := reader_testing.NewSimpleReader(log, url, name, "example_type", time.Hour, time.Hour, 10)
		red.ReadFunc = readFunc(red)
		reds[i] = red
	}
	return reds
}
