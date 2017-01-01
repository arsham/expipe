// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/recorder"
)

// testRecorderErrorsOnUnavailableEndpoint tests the recorder errors for bad URL.
func testRecorderErrorsOnUnavailableEndpoint(t *testing.T, rec recorder.DataRecorder, err error) {
	if err == nil {
		t.Error("want error, got nil")
	}
	if _, ok := err.(interface {
		EndpointNotAvailable()
	}); !ok {
		t.Errorf("want EndpointNotAvailable, got (%v)", err)
	}
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("want (nil), got (%v)", rec)
	}
}

// testRecorderBacksOffOnEndpointGone is a helper to test the recorder backs off when the endpoint goes away.
func testRecorderBacksOffOnEndpointGone(t *testing.T, rec recorder.DataRecorder, teardown func()) {
	ctx := context.Background()
	teardown()
	p := datatype.NewContainer([]datatype.DataType{})
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
