// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expipe_test

import (
	"context"
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
		io.WriteString(w, desire)
	}))

	red, err := reader_testing.New(log, ts.URL, "reader_example", "typeName", time.Millisecond*100, time.Millisecond*100, 5) //for testing
	if err != nil {
		panic(err)
	}
	red.Pinged = true
	return map[string]reader.DataReader{red.Name(): red}, func() {
		ts.Close()
	}
}

func getRecorder(ctx context.Context, log internal.FieldLogger, url string) recorder.DataRecorder {
	rec, err := recorder_testing.New(ctx, log, "reader_example", url, "intexName", time.Millisecond*100, 5)
	if err != nil {
		panic(err)
	}
	rec.Pinged = true
	return rec
}
