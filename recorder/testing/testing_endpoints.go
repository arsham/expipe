// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/recorder"
)

// testRecorderErrorsOnUnavailableEndpoint tests the recorder errors for bad URL.
func testRecorderErrorsOnUnavailableEndpoint(t *testing.T, cons Constructor) {
	timeout := 10 * time.Millisecond
	name := "the name"
	indexName := "my_index_name"
	backoff := 5
	cons.SetName(name)
	cons.SetIndexName(indexName)
	cons.SetEndpoint(cons.ValidEndpoints()[0])
	cons.SetTimeout(timeout)
	cons.SetBackoff(backoff)

	rec, err := cons.Object()
	if err != nil {
		t.Fatalf("want nil, got (%v)", err)
	}

	err = rec.Ping()
	if err == nil {
		t.Error("want error, got nil")
	}
	if _, ok := err.(interface {
		EndpointNotAvailable()
	}); !ok {
		t.Errorf("want EndpointNotAvailable, got (%v)", err)
	}
}

// testRecorderBacksOffOnEndpointGone is a helper to test the recorder backs off when the endpoint goes away.
func testRecorderBacksOffOnEndpointGone(t *testing.T, cons Constructor) {
	ctx := context.Background()
	ts := cons.TestServer()
	timeout := 10 * time.Millisecond
	cons.SetName("the name")
	cons.SetIndexName("my_index_name")
	cons.SetEndpoint(ts.URL)
	cons.SetTimeout(timeout)
	cons.SetBackoff(5)

	rec, err := cons.Object()
	if err != nil {
		t.Fatal(err)
	}
	err = rec.Ping()
	if err != nil {
		t.Fatal(err)
	}
	ts.Close()
	p := datatype.New([]datatype.DataType{})
	payload := &recorder.RecordJob{
		ID:        communication.NewJobID(),
		Payload:   p,
		IndexName: "my index",
		TypeName:  "my type",
		Time:      time.Now(),
	}

	// We don't know the backoff amount set in the recorder, so we try 100 times until it closes.
	backedOff := false
	for i := 0; i < 100; i++ {
		err := rec.Record(ctx, payload)
		if err == recorder.ErrBackoffExceeded {
			backedOff = true
			break
		}
	}
	if !backedOff {
		t.Error("expected to receive a ErrBackoffExceeded")
	}

	// sending another job, it should block
	done := make(chan struct{})
	go func() {
		rec.Record(ctx, payload)
		close(done)
	}()
	select {
	case <-done:
		// good one!
	case <-time.After(20 * time.Millisecond):
		t.Error("expected the recorder to be gone")
	}
}

// testRecordingReturnsErrorIfNotPingedYet is a helper to test the recorder returns an error
// if the caller hasn't called the Ping() method.
func testRecordingReturnsErrorIfNotPingedYet(t *testing.T, cons Constructor) {
	ctx := context.Background()
	timeout := 10 * time.Millisecond
	cons.SetName("the name")
	cons.SetIndexName("my_index_name")
	cons.SetTimeout(timeout)
	cons.SetBackoff(5)

	rec, err := cons.Object()
	if err != nil {
		t.Fatal(err)
	}
	p := datatype.New([]datatype.DataType{})
	payload := &recorder.RecordJob{
		ID:        communication.NewJobID(),
		Payload:   p,
		IndexName: "my index",
		TypeName:  "my type",
		Time:      time.Now(),
	}

	if rec.Record(ctx, payload) != recorder.ErrPingNotCalled {
		t.Errorf("want ErrHasntCalledPing, got (%v)", err)
	}
}
