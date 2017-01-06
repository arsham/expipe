// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	reader_testing "github.com/arsham/expvastic/reader/testing"
	"github.com/arsham/expvastic/recorder"
	recorder_testing "github.com/arsham/expvastic/recorder/testing"
	"github.com/arsham/expvastic/token"
)

// TODO: test engine closes readers when recorder goes out of scope

var (
	log        logrus.FieldLogger
	testServer *httptest.Server
)

func TestMain(m *testing.M) {
	log = lib.DiscardLogger()
	testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	exitCode := m.Run()
	testServer.Close()
	os.Exit(exitCode)
}

func TestNewWithReadRecorder(t *testing.T) {
	ctx := context.Background()

	rec, err := recorder_testing.New(ctx, log, "a", testServer.URL, "aa", time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	red, err := reader_testing.New(log, testServer.URL, "a", "dd", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	red2, err := reader_testing.New(log, testServer.URL, "a", "dd", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}

	e, err := expvastic.New(ctx, log, rec, red, red2)
	if err != expvastic.ErrDuplicateRecorderName {
		t.Error("want error, got nil")
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}
}

func TestEngineSendJob(t *testing.T) {
	var recorderID token.ID
	ctx, cancel := context.WithCancel(context.Background())

	red, err := reader_testing.New(log, testServer.URL, "reader_example", "example_type", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	red.ReadFunc = func(job *token.Context) (*reader.Result, error) {
		recorderID = token.NewUID()
		resp := &reader.Result{
			ID:       recorderID,
			Content:  []byte(`{"devil":666}`),
			TypeName: red.TypeName(),
			Mapper:   red.Mapper(),
		}
		return resp, nil
	}

	rec, err := recorder_testing.New(ctx, log, "recorder_example", testServer.URL, "intexName", time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if job.ID != recorderID {
			t.Errorf("want (%d), got (%s)", recorderID, job.ID)
		}
		return nil
	}

	e, err := expvastic.New(ctx, log, rec, red)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}
}

func TestEngineMultiReader(t *testing.T) {
	count := 10
	ctx, cancel := context.WithCancel(context.Background())
	IDs := make([]string, count)
	idChan := make(chan token.ID)
	for i := 0; i < count; i++ {
		id := token.NewUID()
		IDs[i] = id.String()
		go func(id token.ID) {
			idChan <- id
		}(id)
	}

	rec, err := recorder_testing.New(ctx, log, "recorder_example", testServer.URL, "intexName", time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	rec.RecordFunc = func(ctx context.Context, job *recorder.Job) error {
		if !lib.StringInSlice(job.ID.String(), IDs) {
			t.Errorf("want once of (%s), got (%s)", strings.Join(IDs, ","), job.ID)
		}
		return nil
	}

	reds := make([]reader.DataReader, count)
	for i := 0; i < count; i++ {

		name := fmt.Sprintf("reader_example_%d", i)
		red, err := reader_testing.New(log, testServer.URL, name, "example_type", time.Hour, time.Hour, 5)
		if err != nil {
			t.Fatal(err)
		}
		red.ReadFunc = func(job *token.Context) (*reader.Result, error) {
			resp := &reader.Result{
				ID:       <-idChan,
				Content:  []byte(`{"devil":666}`),
				TypeName: red.TypeName(),
				Mapper:   red.Mapper(),
			}
			return resp, nil
		}
		reds[i] = red
	}

	e, err := expvastic.New(ctx, log, rec, reds...)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}
}

func TestEngineNewWithConfig(t *testing.T) {
	ctx := context.Background()

	red, err := reader_testing.NewConfig("", "reader_example", log, "nowhere", "/still/nowhere", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	rec, err := recorder_testing.NewConfig("recorder_example", log, "nowhere", time.Hour, 5, "index")
	if err != nil {
		t.Fatal(err)
	}

	e, err := expvastic.WithConfig(ctx, log, rec, red)
	if err != reader.ErrEmptyName {
		t.Errorf("want ErrEmptyReaderName, got (%v)", err)
	}
	if e != nil {
		t.Errorf("want nil, got (%v)", e)
	}

	// triggering recorder errors
	rec, _ = recorder_testing.NewConfig("recorder_example", log, "nowhere", time.Hour, 5, "index")
	red, _ = reader_testing.NewConfig("same_name_is_illegal", "reader_example", log, testServer.URL, "/still/nowhere", time.Hour, time.Hour, 5)

	e, err = expvastic.WithConfig(ctx, log, rec, red)
	if e != nil {
		t.Errorf("want nil, got (%v)", e)
	}
	if _, ok := err.(interface {
		InvalidEndpoint()
	}); !ok {
		t.Errorf("want ErrInvalidEndpoint, got (%v)", err)
	}

	red, _ = reader_testing.NewConfig("same_name_is_illegal", "reader_example", log, testServer.URL, "/still/nowhere", time.Hour, time.Hour, 5)
	red2, _ := reader_testing.NewConfig("same_name_is_illegal", "reader_example", log, testServer.URL, "/still/nowhere", time.Hour, time.Hour, 5)
	rec, _ = recorder_testing.NewConfig("recorder_example", log, testServer.URL, time.Hour, 5, "index")
	e, err = expvastic.WithConfig(ctx, log, rec, red, red2)
	if err == nil {
		t.Error("want error, got nil")
	}
	if err != expvastic.ErrDuplicateRecorderName {
		t.Errorf("want ErrDuplicateRecorderName, got (%v)", err)
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}
}

func TestEngineErrorsIfReaderNotPinged(t *testing.T) {
	ctx := context.Background()
	redServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	recServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer recServer.Close()
	redServer.Close() // making sure no one else is got this random port at this time

	rec, err := recorder_testing.New(ctx, log, "a", recServer.URL, "aa", time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	red, err := reader_testing.New(log, redServer.URL, "a", "dd", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}

	e, err := expvastic.New(ctx, log, rec, red)
	if err == nil {
		t.Error("want ErrPing, got nil")
	}

	if _, ok := err.(interface {
		Ping()
	}); !ok {
		t.Errorf("want ErrPing, got (%v)", err)
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}
}

func TestEngineErrorsIfRecorderNotPinged(t *testing.T) {
	ctx := context.Background()
	redServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	recServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	recServer.Close() // making sure no one else is got this random port at this time
	defer redServer.Close()

	rec, err := recorder_testing.New(ctx, log, "a", recServer.URL, "aa", time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	red, err := reader_testing.New(log, redServer.URL, "a", "dd", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}

	e, err := expvastic.New(ctx, log, rec, red)
	if err == nil {
		t.Error("want ErrPing, got nil")
	}

	if _, ok := err.(interface {
		Ping()
	}); !ok {
		t.Errorf("want ErrPing, got (%v)", err)
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}
}
