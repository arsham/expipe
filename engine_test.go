// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	reader_testing "github.com/arsham/expvastic/reader/testing"
	"github.com/arsham/expvastic/recorder"
	recorder_testing "github.com/arsham/expvastic/recorder/testing"
)

// TODO: test engine closes readers when recorder goes out of scope

func TestNewWithReadRecorder(t *testing.T) {
	log := lib.DiscardLogger()
	ctx := context.Background()

	rec, err := recorder_testing.NewSimpleRecorder(ctx, log, "a", "http://127.0.0.1:9200", "aa", time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	red, err := reader_testing.NewSimpleReader(log, "http://127.0.0.1:9200", "a", "dd", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	red2, err := reader_testing.NewSimpleReader(log, "http://127.0.0.1:9200", "a", "dd", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}

	e, err := expvastic.NewWithReadRecorder(ctx, log, rec, red, red2)
	if err != expvastic.ErrDuplicateRecorderName {
		t.Error("want error, got nil")
	}
	if e != nil {
		t.Errorf("want (nil), got (%v)", e)
	}
}

func TestEngineSendJob(t *testing.T) {
	var recorderID communication.JobID
	log := lib.DiscardLogger()
	ctx, cancel := context.WithCancel(context.Background())

	red, err := reader_testing.NewSimpleReader(log, "http://127.0.0.1:9200", "reader_example", "example_type", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	red.ReadFunc = func(ctx context.Context) (*reader.ReadJobResult, error) {
		recorderID = communication.NewJobID()
		resp := &reader.ReadJobResult{
			ID:       recorderID,
			Res:      []byte(`{"devil":666}`),
			TypeName: red.TypeName(),
			Mapper:   red.Mapper(),
		}
		return resp, nil
	}

	rec, err := recorder_testing.NewSimpleRecorder(ctx, log, "recorder_example", "http://127.0.0.1:9200", "intexName", time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	rec.RecordFunc = func(ctx context.Context, job *recorder.RecordJob) error {
		if job.ID != recorderID {
			t.Errorf("want (%d), got (%s)", recorderID, job.ID)
		}
		return nil
	}

	e, err := expvastic.NewWithReadRecorder(ctx, log, rec, red)
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
	log := lib.DiscardLogger()
	ctx, cancel := context.WithCancel(context.Background())
	IDs := make([]string, count)
	idChan := make(chan communication.JobID)
	for i := 0; i < count; i++ {
		id := communication.NewJobID()
		IDs[i] = id.String()
		go func(id communication.JobID) {
			idChan <- id
		}(id)
	}

	rec, err := recorder_testing.NewSimpleRecorder(ctx, log, "recorder_example", "http://127.0.0.1:9200", "intexName", time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	rec.RecordFunc = func(ctx context.Context, job *recorder.RecordJob) error {
		if !lib.StringInSlice(job.ID.String(), IDs) {
			t.Errorf("want once of (%s), got (%s)", strings.Join(IDs, ","), job.ID)
		}
		return nil
	}

	reds := make([]reader.DataReader, count)
	for i := 0; i < count; i++ {

		name := fmt.Sprintf("reader_example_%d", i)
		red, err := reader_testing.NewSimpleReader(log, "http://127.0.0.1:9200", name, "example_type", time.Hour, time.Hour, 5)
		if err != nil {
			t.Fatal(err)
		}
		red.ReadFunc = func(ctx context.Context) (*reader.ReadJobResult, error) {
			resp := &reader.ReadJobResult{
				ID:       <-idChan,
				Res:      []byte(`{"devil":666}`),
				TypeName: red.TypeName(),
				Mapper:   red.Mapper(),
			}
			return resp, nil
		}
		reds[i] = red
	}

	e, err := expvastic.NewWithReadRecorder(ctx, log, rec, reds...)
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
	log := lib.DiscardLogger()

	red, err := reader_testing.NewMockConfig("", "reader_example", log, "nowhere", "/still/nowhere", time.Hour, time.Hour, 5)
	if err != nil {
		t.Fatal(err)
	}
	rec, err := recorder_testing.NewMockConfig("recorder_example", log, "nowhere", time.Hour, 5, "index")
	if err != nil {
		t.Fatal(err)
	}

	e, err := expvastic.NewWithConfig(ctx, log, rec, red)
	if err != reader.ErrEmptyName {
		t.Errorf("want ErrEmptyReaderName, got (%v)", err)
	}
	if e != nil {
		t.Errorf("want nil, got (%v)", e)
	}

	// triggering recorder errors
	rec, _ = recorder_testing.NewMockConfig("recorder_example", log, "nowhere", time.Hour, 5, "index")
	red, _ = reader_testing.NewMockConfig("same_name_is_illegal", "reader_example", log, "http://127.0.0.1:9200", "/still/nowhere", time.Hour, time.Hour, 5)

	e, err = expvastic.NewWithConfig(ctx, log, rec, red)
	if e != nil {
		t.Errorf("want nil, got (%v)", e)
	}
	if _, ok := err.(interface {
		InvalidEndpoint()
	}); !ok {
		t.Errorf("want ErrInvalidEndpoint, got (%v)", err)
	}

	red, _ = reader_testing.NewMockConfig("same_name_is_illegal", "reader_example", log, "http://127.0.0.1:9200", "/still/nowhere", time.Hour, time.Hour, 5)
	red2, _ := reader_testing.NewMockConfig("same_name_is_illegal", "reader_example", log, "http://127.0.0.1:9200", "/still/nowhere", time.Hour, time.Hour, 5)
	rec, _ = recorder_testing.NewMockConfig("recorder_example", log, "http://127.0.0.1:9200", time.Hour, 5, "index")
	e, err = expvastic.NewWithConfig(ctx, log, rec, red, red2)
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
