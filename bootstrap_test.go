// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expipe_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expipe"
	"github.com/arsham/expipe/config"
	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/reader"
	rdt "github.com/arsham/expipe/reader/testing"
	"github.com/arsham/expipe/recorder"
	rct "github.com/arsham/expipe/recorder/testing"
)

func requirements(t *testing.T) (context.Context, *internal.Logger, *rdt.Reader, *rct.Recorder) {
	log := internal.DiscardLogger()
	ctx := context.Background()
	mockReader, err := rdt.New(
		reader.WithName("name"),
		reader.WithEndpoint("127.0.0.1"),
		reader.WithTypeName("s"),
		reader.WithBackoff(5),
		reader.WithInterval(time.Second),
		reader.WithTimeout(time.Second),
		reader.WithLogger(log),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil): getting requirements for reader", err)
	}
	mockRecorder, err := rct.New(
		recorder.WithName("name"),
		recorder.WithEndpoint("127.0.0.1"),
		recorder.WithIndexName("in"),
		recorder.WithBackoff(5),
		recorder.WithTimeout(time.Second),
		recorder.WithLogger(log),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil): getting requirements for recorder", err)
	}
	return ctx, log, mockReader, mockRecorder
}

func TestEmptyConfmapErrors(t *testing.T) {
	t.Parallel()
	ctx, log, _, _ := requirements(t)
	d, err := expipe.StartEngines(ctx, log, nil)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if d != nil {
		t.Errorf("d = (%v); want (nil)", d)
	}
}

func TestEmptyReaderErrors(t *testing.T) {
	t.Parallel()
	ctx, log, _, mockRecorder := requirements(t)
	confMap := &config.ConfMap{
		Readers:   map[string]reader.DataReader{},
		Recorders: map[string]recorder.DataRecorder{"rec1": mockRecorder},
		Routes:    map[string][]string{"rec1": {"red1", "red2"}},
	}
	d, err := expipe.StartEngines(ctx, log, confMap)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if d != nil {
		t.Errorf("d = (%v); want (nil)", d)
	}
}

func TestEmptyRecorderErrors(t *testing.T) {
	t.Parallel()
	ctx, log, mockReader, _ := requirements(t)
	confMap := &config.ConfMap{
		Readers:   map[string]reader.DataReader{"red1": mockReader},
		Recorders: map[string]recorder.DataRecorder{"rec1": nil},
		Routes:    map[string][]string{"rec1": {"red1", "red2"}},
	}
	d, err := expipe.StartEngines(ctx, log, confMap)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if d != nil {
		t.Errorf("d = (%v); want (nil)", d)
	}
}

func TestEmptyReaderNameErrors(t *testing.T) {
	t.Parallel()
	ctx, log, mockReader, mockRecorder := requirements(t)
	mockReader.SetName("")
	confMap := &config.ConfMap{
		Readers:   map[string]reader.DataReader{"red1": mockReader},
		Recorders: map[string]recorder.DataRecorder{"rec1": mockRecorder},
		Routes:    map[string][]string{"rec1": {"red1", "red2"}},
	}
	d, err := expipe.StartEngines(ctx, log, confMap)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if d != nil {
		t.Errorf("d = (%v); want (nil)", d)
	}
}

func TestEmptyRecorderNameErrors(t *testing.T) {
	t.Parallel()
	var (
		ctx context.Context
		log *internal.Logger
		rec *rct.Recorder
		red *rdt.Reader
	)
	ctx, log, red, rec = requirements(t)
	rec.SetName("")
	confMap := &config.ConfMap{
		Readers: map[string]reader.DataReader{
			"red1": red,
		},
		Recorders: map[string]recorder.DataRecorder{
			"rec1": rec,
		},
		Routes: map[string][]string{"rec1": {"red1", "red2"}},
	}
	d, err := expipe.StartEngines(ctx, log, confMap)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if d != nil {
		t.Errorf("d = (%v); want (nil)", d)
	}
}

func TestClosesDoneChan(t *testing.T) {
	t.Parallel()
	var (
		ctx context.Context
		log *internal.Logger
		red *rdt.Reader
		rec *rct.Recorder
	)
	ctx, log, red, rec = requirements(t)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer testServer.Close()
	rec.SetEndpoint(testServer.URL)
	red.SetEndpoint(testServer.URL)

	confMap := &config.ConfMap{
		Readers: map[string]reader.DataReader{
			"red1": red,
		},
		Recorders: map[string]recorder.DataRecorder{
			"rec1": rec,
		},
		Routes: map[string][]string{"rec1": {"red1"}},
	}
	ctx, cancel := context.WithCancel(ctx)
	d, err := expipe.StartEngines(ctx, log, confMap)
	if err != nil {
		t.Fatalf("err = (%v); want (nil)", err)
	}
	if d == nil {
		t.Fatal("d = (nil); want (chan struct{})")
	}

	select {
	case <-d:
		t.Error("Expected not to stop")
	case <-time.After(time.Millisecond * 100):
	}

	cancel()
	select {
	case <-d:
	case <-time.After(time.Second * 5):
		t.Error("Expected close the done channel")
	}
}
