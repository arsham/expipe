// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/reader"
	reader_testing "github.com/arsham/expvastic/reader/testing"
	"github.com/arsham/expvastic/recorder"
	recorder_testing "github.com/arsham/expvastic/recorder/testing"
)

func getReader(log logrus.FieldLogger) (reader.DataReader, func()) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		desire := `{"the key": "is the value!"}`
		io.WriteString(w, desire)
	}))

	red, err := reader_testing.NewSimpleReader(log, ts.URL, "reader_example", "typeName", time.Second, time.Second, 5) //for testing
	if err != nil {
		panic(err)
	}
	return red, func() {
		ts.Close()
	}
}

func getRecorder(ctx context.Context, log logrus.FieldLogger, url string) recorder.DataRecorder {
	rec, err := recorder_testing.NewSimpleRecorder(ctx, log, "reader_example", url, "intexName", time.Second, 5)
	if err != nil {
		panic(err)
	}
	return rec
}
