// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"testing"
	"time"

	"github.com/arsham/expipe/internal/datatype"
	"github.com/arsham/expipe/internal/token"
	"github.com/arsham/expipe/recorder"
)

// testRecorderReceivesPayload tests the recorder receives the payload correctly.
func testRecorderReceivesPayload(t *testing.T, cons Constructor) {
	ctx := context.Background()
	cons.SetName("the name")
	cons.SetIndexName("my_index")
	cons.SetTimeout(time.Second)
	cons.SetBackoff(5)
	cons.SetEndpoint(cons.TestServer().URL)

	rec, err := cons.Object()
	if err != nil {
		t.Fatal(err)
	}
	rec.Ping()
	p := datatype.New([]datatype.DataType{})
	payload := &recorder.Job{
		ID:        token.NewUID(),
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
func testRecorderSendsResult(t *testing.T, cons Constructor) {
	ctx := context.Background()
	cons.SetName("the name")
	cons.SetIndexName("index_name")
	cons.SetTimeout(time.Second)
	cons.SetBackoff(15)
	cons.SetEndpoint(cons.TestServer().URL)

	rec, err := cons.Object()
	if err != nil {
		t.Fatal(err)
	}
	err = rec.Ping()
	if err != nil {
		t.Fatal(err)
	}
	p := datatype.New([]datatype.DataType{&datatype.StringType{Key: "test", Value: "test"}})
	payload := &recorder.Job{
		ID:        token.NewUID(),
		Payload:   p,
		IndexName: "my_index",
		TypeName:  "my_type",
		Time:      time.Now(),
	}

	err = rec.Record(ctx, payload)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}
