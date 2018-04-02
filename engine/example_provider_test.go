// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"time"

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
