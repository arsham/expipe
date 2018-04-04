// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/arsham/expipe/engine"
	"github.com/arsham/expipe/reader"
	rdt "github.com/arsham/expipe/reader/testing"
	"github.com/arsham/expipe/recorder"
	rct "github.com/arsham/expipe/recorder/testing"
	"github.com/arsham/expipe/tools"
)

func getReader(log tools.FieldLogger) reader.DataReader {
	done := make(chan struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		desire := `{"the key": "is the value!"}`
		_, err := io.WriteString(w, desire)
		if err != nil {
			panic(err)
		}
		close(done)
	}))

	go func() {
		<-done
		ts.Close()
	}()

	red, err := rdt.New(
		reader.WithLogger(log),
		reader.WithEndpoint(ts.URL),
		reader.WithName("reader_example"),
		reader.WithTypeName("typeName"),
		reader.WithInterval(time.Millisecond*100),
		reader.WithTimeout(time.Second),
		reader.WithBackoff(5),
	)

	if err != nil {
		panic(err)
	}
	red.Pinged = true
	return red
}

func getRecorders(log tools.FieldLogger, url string) map[string]recorder.DataRecorder {
	rec := getRecorder(log, url)
	return map[string]recorder.DataRecorder{rec.Name(): rec}
}

func getRecorder(log tools.FieldLogger, url string) recorder.DataRecorder {
	rec, err := rct.New(
		recorder.WithLogger(log),
		recorder.WithEndpoint(url),
		recorder.WithName("recorder_example"),
		recorder.WithIndexName("indexName"),
		recorder.WithTimeout(time.Second),
		recorder.WithBackoff(5),
	)
	if err != nil {
		panic(err)
	}
	rec.Pinged = true
	return rec
}

// engineWithReadRecs creates an Engine instance with already set-up readers and
// recorders. The Engine's work starts from here by streaming reader payloads
// to the recorders. Returns an error if there are recorders with the same
// name, or any of constructions results in errors.
func engineWithReadRecs(ctx context.Context, log tools.FieldLogger, red reader.DataReader, recs map[string]recorder.DataRecorder) (engine.Engine, error) {
	failedErrors := make(map[string]error)
	err := red.Ping()
	if err != nil {
		return nil, engine.PingError{red.Name(): err}
	}

	recorders := make([]recorder.DataRecorder, 0)
	canDo := false
	for name, rec := range recs {
		err := rec.Ping()
		if err != nil {
			failedErrors[name] = err
			continue
		}
		recorders = append(recorders, rec)
		canDo = true
	}
	if !canDo {
		return nil, engine.PingError(failedErrors)
	}
	return engine.New(
		engine.WithCtx(ctx),
		engine.WithLogger(log),
		engine.WithReader(red),
		engine.WithRecorders(recorders...),
	)
}
