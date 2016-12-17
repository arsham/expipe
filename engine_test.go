// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/lib"
)

func sampleSetup(log *logrus.Logger, url string, reader expvastic.TargetReader, rec *mockRecorder) *expvastic.Conf {
	return &expvastic.Conf{
		Logger:       log,
		TargetReader: reader,
		Recorder:     rec,
		Target:       url,
		Interval:     1 * time.Millisecond, Timeout: 3 * time.Millisecond,
	}
}

func TestEngineSendsTheJob(t *testing.T) {
	log := lib.DiscardLogger()
	bg := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	rec := &mockRecorder{
		PayloadChanFunc: func() chan *expvastic.RecordJob {
			return nil
		},
	}
	jobSent := make(chan struct{})
	read := &MockCtxReader{
		ContextReadFunc: func(ctx context.Context) (*http.Response, error) {
			jobSent <- struct{}{}
			return http.Get(ts.URL)
		},
	}
	reader, _ := expvastic.NewExpvarReader(log, read)
	reader.Start()
	conf := sampleSetup(log, ts.URL, reader, rec)

	ctx, cancel := context.WithCancel(bg)
	cl := expvastic.NewEngine(ctx, *conf)
	go cl.Start()

	select {
	case <-ctx.Done():
		t.Error("job wasn't sent")
	case <-jobSent:
		cancel()
	}

	<-ctx.Done()
}

func TestEngineRecorderReturnsCorrectResult(t *testing.T) {
	bg := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()
	ftype := expvastic.FloatType{"test", 666.66}
	ftypeStr := fmt.Sprintf("{%s}", ftype)
	resultChan := make(chan *expvastic.RecordJob)
	rec := &mockRecorder{
		PayloadChanFunc: func() chan *expvastic.RecordJob {
			return resultChan
		},
	}
	resCh := make(chan expvastic.ReadJobResult)
	jobs := make(chan context.Context)
	reader := NewMockExpvarReader(jobs, resCh, func(c context.Context) {})
	log := lib.DiscardLogger()
	conf := sampleSetup(log, ts.URL, reader, rec)
	ctx, _ := context.WithTimeout(bg, 5*time.Millisecond)
	cl := expvastic.NewEngine(ctx, *conf)
	defer cl.Stop()
	buf := new(bytes.Buffer)
	log.Out = buf
	go cl.Start()
	res := reader.ResultChan()
	msg := ioutil.NopCloser(strings.NewReader(ftypeStr))
	res <- expvastic.ReadJobResult{Res: msg}
	r := <-resultChan
	result := r.Payload.List()[0]
	if result.String() != ftype.String() {
		t.Errorf("want (%s), got (%s)", ftype.String(), result.String())
	}
	<-ctx.Done()
}

func TestEngineStop(t *testing.T) {
	t.Skip("Not implemented here")
}
