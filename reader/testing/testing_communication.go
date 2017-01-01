// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/reader"
)

// testReaderReceivesJob is a test helper to test the reader can receive jobs
func testReaderReceivesJob(t *testing.T, red reader.DataReader) {
	ctx := context.Background()
	done := make(chan struct{})
	go func() {
		result, err := red.Read(communication.NewReadJob(ctx))
		if err != nil {
			t.Errorf("want nil, got (%v)", err)
		}
		if result == nil {
			t.Fatal("expecting results, got nil")
		}
		if result.ID.String() == "" {
			t.Error("expecting ID, got nil")
		}
		if result.TypeName == "" {
			t.Error("expecting TypeName, got empty string")
		}
		if result.Res == nil {
			t.Error("expecting Res, got nil")
		}
		if result.Mapper == nil {
			t.Error("expecting Mapper, got nil")
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected the reader to receive the job, but it blocked")
	}
}

// testReaderReturnsSameID is a test helper to test the reader returns the same ID in the response
func testReaderReturnsSameID(t *testing.T, red reader.DataReader) {
	done := make(chan struct{})
	go func() {
		ctx := context.Background()
		job := communication.NewReadJob(ctx)
		result, err := red.Read(job)
		if err != nil {
			t.Errorf("want nil, got (%v)", err)
		}
		if result == nil {
			t.Fatal("expecting results, got nil")
		}
		if result.ID != communication.JobValue(job) {
			t.Errorf("want (%v), got (%v)", communication.JobValue(job), result.ID)
		}

		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected the reader to receive the job, but it blocked")
	}
}
