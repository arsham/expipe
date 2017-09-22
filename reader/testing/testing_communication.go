// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/arsham/expvastic/internal/token"
)

// testReaderReceivesJob is a test helper to test the reader can receive jobs
func testReaderReceivesJob(t *testing.T, cons Constructor) {
	cons.SetName("the name")
	cons.SetTypename("my type")
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetInterval(time.Hour)
	cons.SetTimeout(time.Hour)
	cons.SetBackoff(5)

	red, err := cons.Object()
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}
	err = red.Ping()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	done := make(chan struct{})
	errChan := make(chan string)
	fatalChan := make(chan string)
	go func() {
		result, err := red.Read(token.New(ctx))
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
		if result.Content == nil {
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
func testReaderReturnsSameID(t *testing.T, cons Constructor) {
	cons.SetName("the name")
	cons.SetTypename("my type")
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetInterval(time.Hour)
	cons.SetTimeout(time.Hour)
	cons.SetBackoff(5)
	red, err := cons.Object()
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}
	err = red.Ping()
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	errChan := make(chan string)
	fatalChan := make(chan string)
	go func() {
		ctx := context.Background()
		job := token.New(ctx)
		result, err := red.Read(job)
		if err != nil {
			errChan <- fmt.Sprintf("want nil, got (%v)", err)
		}
		if result == nil {
			fatalChan <- "expecting results, got nil"
		}
		if result.ID != job.ID() {
			errChan <- fmt.Sprintf("want (%v), got (%v)", job.ID(), result.ID)
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
