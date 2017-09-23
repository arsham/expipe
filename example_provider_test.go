// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expipe_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/reader"
	reader_testing "github.com/arsham/expipe/reader/testing"
	"github.com/arsham/expipe/recorder"
	recorder_testing "github.com/arsham/expipe/recorder/testing"
)

func getReader(log internal.FieldLogger) (map[string]reader.DataReader, func()) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		desire := `{"the key": "is the value!"}`
		_, err := io.WriteString(w, desire)
		if err != nil {
			panic(err)
		}
	}))

	red, err := reader_testing.New(
		reader.SetLogger(log),
		reader.SetEndpoint(ts.URL),
		reader.SetName("reader_example"),
		reader.SetTypeName("typeName"),
		reader.SetInterval(time.Millisecond*100),
		reader.SetTimeout(time.Second),
		reader.SetBackoff(5),
	)

	if err != nil {
		panic(err)
	}
	red.Pinged = true
	return map[string]reader.DataReader{red.Name(): red}, func() {
		ts.Close()
	}
}

func getRecorder(log internal.FieldLogger, url string) recorder.DataRecorder {
	rec, err := recorder_testing.New(
		recorder.SetLogger(log),
		recorder.SetEndpoint(url),
		recorder.SetName("recorder_example"),
		recorder.SetIndexName("indexName"),
		recorder.SetTimeout(time.Second),
		recorder.SetBackoff(5),
	)
	if err != nil {
		panic(err)
	}
	rec.Pinged = true
	return rec
}
