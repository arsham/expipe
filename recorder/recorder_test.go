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

    "github.com/arsham/expvastic/lib"
)

// The purpose of these tests is to make sure the simple recorder, which is a mock,
// works perfect, so other tests can rely on it.

func TestSimpleRecorderReceivesPayload(t *testing.T) {
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
    defer ts.Close()

    payloadChan := make(chan *RecordJob)
    rec, _ := NewSimpleRecorder(ctx, log, payloadChan, "reader_example", ts.URL, "intexName", 10*time.Millisecond)
    rec.Start(ctx)

    errChan := make(chan error)
    payload := &RecordJob{
        Ctx:       ctx,
        Payload:   nil,
        IndexName: "my index",
        Time:      time.Now(),
        Err:       errChan,
    }
    select {
    case rec.PayloadChan() <- payload:
    case <-time.After(5 * time.Second):
        t.Error("expected the recorder to recive the payload, but it blocked")
    }
}

func TestSimpleRecorderSendsResult(t *testing.T) {
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
    defer ts.Close()

    payloadChan := make(chan *RecordJob)
    rec, _ := NewSimpleRecorder(ctx, log, payloadChan, "reader_example", ts.URL, "intexName", 10*time.Millisecond)
    rec.Start(ctx)

    errChan := make(chan error)
    payload := &RecordJob{
        Ctx:       ctx,
        Payload:   nil,
        IndexName: "my index",
        Time:      time.Now(),
        Err:       errChan,
    }
    rec.PayloadChan() <- payload

    select {
    case err := <-errChan:
        if err != nil {
            t.Errorf("want (nil), got (%v)", err)
        }
    case <-time.After(5 * time.Second):
        t.Error("expected to recive a data back, nothing recieved")
    }
}

func TestSimpleRecorderErrorsOnBadURL(t *testing.T) {
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    payloadChan := make(chan *RecordJob)
    rec, _ := NewSimpleRecorder(ctx, log, payloadChan, "reader_example", "leads nowhere", "intexName", 10*time.Millisecond)
    rec.Start(ctx)

    errChan := make(chan error)
    payload := &RecordJob{
        Ctx:       ctx,
        Payload:   nil,
        IndexName: "my index",
        Time:      time.Now(),
        Err:       errChan,
    }
    rec.PayloadChan() <- payload

    select {
    case err := <-errChan:
        if err == nil {
            t.Errorf("want (nil), got (%v)", err)
        }
    case <-time.After(5 * time.Second):
        t.Error("expected to recive a data back, nothing recieved")
    }
}

func TestSimpleRecorderCloses(t *testing.T) {
    log := lib.DiscardLogger()
    ctx, cancel := context.WithCancel(context.Background())

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
    defer ts.Close()

    payloadChan := make(chan *RecordJob)
    rec, _ := NewSimpleRecorder(ctx, log, payloadChan, "reader_example", ts.URL, "intexName", 10*time.Millisecond)
    doneChan := rec.Start(ctx)
    select {
    case <-doneChan:
        t.Error("expected the recorder to continue working")
    default:
    }

    cancel()

    select {
    case <-doneChan:
    case <-time.After(5 * time.Second):
        t.Error("expected the recorder to quit working")
    }
}
