// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/recorder"
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
		errorChan := make(chan communication.ErrorMessage, bc.recChanBuff+(bc.readers*bc.readChanBuff))
		payloadChan := make(chan *recorder.RecordJob, bc.recChanBuff)
		resultChan := make(chan *reader.ReadJobResult, bc.readResChanBuff)

		// Setting the intervals to an hour so the benchmark can issue jobs
		rec, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, errorChan, "reacorder_example", "nowhere, it doesn't matter", "intexName", time.Hour)
		reds := makeReaders(ctx, bc.readers, log, resultChan, errorChan, bc.recChanBuff, "nowhere, it doesn't matter")
		e, _ := expvastic.NewWithReadRecorder(ctx, log, errorChan, resultChan, rec, reds...)

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
			red.JobChan() <- communication.NewReadJob(ctx)
		}
	}
}

func makeReaders(ctx context.Context, count int, log logrus.FieldLogger, resultChan chan *reader.ReadJobResult, errorChan chan communication.ErrorMessage, chanBuff int, url string) []reader.DataReader {
	reds := make([]reader.DataReader, count)
	startFunc := func(m *reader.SimpleReader) func(communication.StopChannel) {
		return func(stop communication.StopChannel) {
			go func() {
				for {
					select {
					case job := <-m.JobChan():
						id := communication.JobValue(job)
						res := &reader.ReadJobResult{
							ID:       id,
							Time:     time.Now(),
							Res:      ioutil.NopCloser(bytes.NewBuffer([]byte(``))),
							TypeName: m.TypeName(),
							Mapper:   m.Mapper(),
						}
						m.ResultChan() <- res
					case s := <-stop:
						s <- struct{}{}
						return
					}
				}
			}()
		}
	}
	for i := 0; i < count; i++ {
		jobChan := make(chan context.Context, chanBuff)
		name := fmt.Sprintf("reader_%d", i)
		red, _ := reader.NewSimpleReader(log, reader.NewCtxReader(url), jobChan, resultChan, errorChan, name, "example_type", time.Hour, time.Hour)
		red.StartFunc = startFunc(red)
		reds[i] = red
	}
	return reds
}
