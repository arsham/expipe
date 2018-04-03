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

func getReader(log tools.FieldLogger) (map[string]reader.DataReader, func()) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		desire := `{"the key": "is the value!"}`
		_, err := io.WriteString(w, desire)
		if err != nil {
			panic(err)
		}
	}))

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
	return map[string]reader.DataReader{red.Name(): red}, func() {
		ts.Close()
	}
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

// engineWithReadRecs creates an Engine instance with already set-up reader and
// recorders. The Engine's work starts from here by streaming all readers
// payloads to the recorder. Returns an error if there are recorders with
// the same name, or any of constructions results in errors.
func engineWithReadRecs(ctx context.Context, log tools.FieldLogger, rec recorder.DataRecorder, reds map[string]reader.DataReader) (*engine.Engine, error) {
	failedErrors := make(map[string]error)
	err := rec.Ping()
	if err != nil {
		return nil, engine.PingError{rec.Name(): err}
	}

	readers := make([]reader.DataReader, 0)
	canDo := false
	for name, red := range reds {
		err := red.Ping()
		if err != nil {
			failedErrors[name] = err
			continue
		}
		readers = append(readers, red)
		canDo = true
	}
	if !canDo {
		return nil, engine.PingError(failedErrors)
	}
	return engine.New(
		engine.WithCtx(ctx),
		engine.WithLogger(log),
		engine.WithReaders(readers...),
		engine.WithRecorder(rec),
	)
}
