// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/config"
	"github.com/arsham/expvastic/internal"
	reader_testing "github.com/arsham/expvastic/reader/testing"
	recorder_testing "github.com/arsham/expvastic/recorder/testing"
)

func TestStartEngines(t *testing.T) {
	t.Parallel()
	log := internal.DiscardLogger()
	ctx := context.Background()
	mockReader := &reader_testing.Config{
		MockName:     "name",
		MockEndpoint: "127.0.0.1",
		MockTypeName: "s",
		MockBackoff:  5,
		MockInterval: time.Second,
		MockTimeout:  time.Second,
	}
	mockRecorder := &recorder_testing.Config{
		MockName:      "name",
		MockEndpoint:  "127.0.0.1",
		MockIndexName: "in",
		MockBackoff:   5,
		MockTimeout:   time.Second,
	}

	t.Run("Empty ConfMap errors", func(t *testing.T) {
		t.Parallel()
		d, err := expvastic.StartEngines(ctx, log, nil)
		if err == nil {
			t.Error("want error, got nil")
		}
		if d != nil {
			t.Errorf("want (nil), got (%v)", d)
		}
	})

	t.Run("Empty reader errors", func(t *testing.T) {
		t.Parallel()
		confMap := &config.ConfMap{
			Readers:   map[string]config.ReaderConf{},
			Recorders: map[string]config.RecorderConf{"rec1": mockRecorder},
			Routes:    map[string][]string{"rec1": {"red1", "red2"}},
		}
		d, err := expvastic.StartEngines(ctx, log, confMap)
		if err == nil {
			t.Error("want error, got nil")
		}
		if d != nil {
			t.Errorf("want (nil), got (%v)", d)
		}
	})

	t.Run("Empty recorder errors", func(t *testing.T) {
		t.Parallel()
		confMap := &config.ConfMap{
			Readers:   map[string]config.ReaderConf{"red1": mockReader},
			Recorders: map[string]config.RecorderConf{"rec1": nil},
			Routes:    map[string][]string{"rec1": {"red1", "red2"}},
		}
		d, err := expvastic.StartEngines(ctx, log, confMap)
		if err == nil {
			t.Error("want error, got nil")
		}
		if d != nil {
			t.Errorf("want (nil), got (%v)", d)
		}
	})

	t.Run("Empty reader name errors", func(t *testing.T) {
		t.Parallel()
		confMap := &config.ConfMap{
			Readers: map[string]config.ReaderConf{
				"red1": &reader_testing.Config{
					MockName: "",
				},
			},
			Recorders: map[string]config.RecorderConf{"rec1": mockRecorder},
			Routes:    map[string][]string{"rec1": {"red1", "red2"}},
		}
		d, err := expvastic.StartEngines(ctx, log, confMap)
		if err == nil {
			t.Error("want error, got nil")
		}
		if d != nil {
			t.Errorf("want (nil), got (%v)", d)
		}
	})

	t.Run("Empty recorder name errors", func(t *testing.T) {
		t.Parallel()
		r := recorder_testing.Config(*mockRecorder)
		r.MockName = ""
		confMap := &config.ConfMap{
			Readers: map[string]config.ReaderConf{
				"red1": mockReader,
			},
			Recorders: map[string]config.RecorderConf{
				"rec1": &r,
			},
			Routes: map[string][]string{"rec1": {"red1", "red2"}},
		}
		d, err := expvastic.StartEngines(ctx, log, confMap)
		if err == nil {
			t.Error("want error, got nil")
		}
		if d != nil {
			t.Errorf("want (nil), got (%v)", d)
		}
	})

	t.Run("Closes done", func(t *testing.T) {
		t.Parallel()
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer testServer.Close()
		rec := recorder_testing.Config(*mockRecorder)
		red := reader_testing.Config(*mockReader)
		rec.MockEndpoint = testServer.URL
		red.MockEndpoint = testServer.URL

		confMap := &config.ConfMap{
			Readers: map[string]config.ReaderConf{
				"red1": &red,
			},
			Recorders: map[string]config.RecorderConf{
				"rec1": &rec,
			},
			Routes: map[string][]string{"rec1": {"red1"}},
		}
		ctx, cancel := context.WithCancel(ctx)
		d, err := expvastic.StartEngines(ctx, log, confMap)
		if err != nil {
			t.Fatalf("want (nil), got (%+v)", err)
		}
		if d == nil {
			t.Fatal("want (chan), got nil")
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

	})
}
