// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/reader"
)

// testReaderReceivesJob is a test helper to test the reader can receive jobs
func testReaderReceivesJob(t *testing.T, red reader.DataReader) {
	ctx := context.Background()
	done := make(chan struct{})
	errChan := make(chan string)
	fatalChan := make(chan string)
	go func() {
		result, err := red.Read(communication.NewReadJob(ctx))
		if err != nil {
			errChan <- fmt.Sprintf("want nil, got (%v)", err)
			return
		}
		if result == nil {
			fatalChan <- "expecting results, got nil"
			return
		}
		if result.ID.String() == "" {
			errChan <- "expecting ID, got nil"
			return
		}
		if result.TypeName == "" {
			errChan <- "expecting TypeName, got empty string"
			return
		}
		if result.Res == nil {
			errChan <- "expecting Res, got nil"
			return
		}
		if result.Mapper == nil {
			errChan <- "expecting Mapper, got nil"
			return
		}
		close(done)
	}()

	select {
	case <-done:
	case msg := <-errChan:
		t.Error(msg)
	case msg := <-fatalChan:
		t.Fatal(msg)
	case <-time.After(5 * time.Second):
		t.Error("expected the reader to receive the job, but it blocked")
	}
}

// testReaderReturnsSameID is a test helper to test the reader returns the same ID in the response
func testReaderReturnsSameID(t *testing.T, red reader.DataReader) {
	done := make(chan struct{})
	errChan := make(chan string)
	fatalChan := make(chan string)
	go func() {
		ctx := context.Background()
		job := communication.NewReadJob(ctx)
		result, err := red.Read(job)
		if err != nil {
			errChan <- fmt.Sprintf("want nil, got (%v)", err)
		}
		if result == nil {
			fatalChan <- "expecting results, got nil"
		}
		if result.ID != communication.JobValue(job) {
			errChan <- fmt.Sprintf("want (%v), got (%v)", communication.JobValue(job), result.ID)
		}

		close(done)
	}()

	select {
	case <-done:
	case msg := <-errChan:
		t.Error(msg)
	case msg := <-fatalChan:
		t.Fatal(msg)
	case <-time.After(5 * time.Second):
		t.Error("expected the reader to receive the job, but it blocked")
	}
}
