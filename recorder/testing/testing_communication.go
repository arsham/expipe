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

// testRecorderReceivesPayload tests the recorder receives the payload correctly.
func testRecorderReceivesPayload(ctx context.Context, t *testing.T, rec recorder.DataRecorder) {
	p := datatype.NewContainer([]datatype.DataType{})
	payload := &recorder.RecordJob{
		ID:        communication.NewJobID(),
		Payload:   p,
		IndexName: "my index",
		TypeName:  "my type",
		Time:      time.Now(),
	}

	received := make(chan struct{})
	go func() {
		rec.Record(ctx, payload)
		received <- struct{}{}
	}()

	select {
	case <-received:
	case <-time.After(5 * time.Second):
		t.Error("expected the recorder to receive the payload, but it blocked")
	}
}

// testRecorderSendsResult tests the recorder send the results to the endpoint.
func testRecorderSendsResult(ctx context.Context, t *testing.T, rec recorder.DataRecorder) {
	p := datatype.NewContainer([]datatype.DataType{&datatype.StringType{Key: "test", Value: "test"}})
	payload := &recorder.RecordJob{
		ID:        communication.NewJobID(),
		Payload:   p,
		IndexName: "my index",
		TypeName:  "my type",
		Time:      time.Now(),
	}

	err := rec.Record(ctx, payload)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}
