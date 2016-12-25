// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/lib"
)

// The purpose of these tests is to make sure the simple recorder, which is a mock,
// works perfect, so other tests can rely on it.

func TestSimpleRecorderReceivesPayload(t *testing.T) {
	t.Parallel()
	log := lib.DiscardLogger()
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	payloadChan := make(chan *RecordJob)
	errorChan := make(chan communication.ErrorMessage)
	rec, _ := NewSimpleRecorder(ctx, log, payloadChan, errorChan, "reader_example", ts.URL, "intexName", 10*time.Millisecond)
	stop := make(communication.StopChannel)
	rec.Start(ctx, stop)

	payload := &RecordJob{
		ID:        communication.NewJobID(),
		Ctx:       ctx,
		Payload:   nil,
		IndexName: "my index",
		Time:      time.Now(),
	}
	select {
	case rec.PayloadChan() <- payload:
	case <-time.After(5 * time.Second):
		t.Error("expected the recorder to receive the payload, but it blocked")
	}
	done := make(chan struct{})
	stop <- done
	<-done

}

func TestSimpleRecorderSendsResult(t *testing.T) {
	t.Parallel()
	log := lib.DiscardLogger()
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	payloadChan := make(chan *RecordJob)
	errorChan := make(chan communication.ErrorMessage)
	rec, _ := NewSimpleRecorder(ctx, log, payloadChan, errorChan, "reader_example", ts.URL, "intexName", 10*time.Millisecond)
	stop := make(communication.StopChannel)
	rec.Start(ctx, stop)

	payload := &RecordJob{
		ID:        communication.NewJobID(),
		Ctx:       ctx,
		Payload:   nil,
		IndexName: "my index",
		Time:      time.Now(),
	}
	rec.PayloadChan() <- payload

	select {
	case err := <-errorChan:
		if err.Err != nil {
			t.Errorf("want (nil), got (%v)", err)
		}
	case <-time.After(20 * time.Millisecond):
	}
	done := make(chan struct{})
	stop <- done
	<-done
}

func TestSimpleRecorderErrorsOnBadURL(t *testing.T) {
	t.Parallel()
	log := lib.DiscardLogger()
	ctx := context.Background()

	payloadChan := make(chan *RecordJob)
	errorChan := make(chan communication.ErrorMessage)
	rec, _ := NewSimpleRecorder(ctx, log, payloadChan, errorChan, "reader_example", "leads nowhere", "intexName", 10*time.Millisecond)
	stop := make(communication.StopChannel)
	rec.Start(ctx, stop)

	payload := &RecordJob{
		ID:        communication.NewJobID(),
		Ctx:       ctx,
		Payload:   nil,
		IndexName: "my index",
		Time:      time.Now(),
	}
	rec.PayloadChan() <- payload

	select {
	case err := <-errorChan:
		if err.Err == nil {
			t.Errorf("want (nil), got (%v)", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("expected to receive a data back, nothing received")
	}
	done := make(chan struct{})
	stop <- done
	<-done
}

func TestSimpleRecorderCloses(t *testing.T) {
	t.Parallel()
	log := lib.DiscardLogger()
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	payloadChan := make(chan *RecordJob)
	errorChan := make(chan communication.ErrorMessage)
	rec, _ := NewSimpleRecorder(ctx, log, payloadChan, errorChan, "reader_example", ts.URL, "intexName", 10*time.Millisecond)
	stop := make(communication.StopChannel)
	rec.Start(ctx, stop)

	done := make(chan struct{})
	stop <- done

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected the recorder to quit working")
	}
}
