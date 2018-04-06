// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"testing"
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/recorder"
	"github.com/arsham/expipe/tools/token"
	"github.com/pkg/errors"
)

// recorderReceivesPayload tests the recorder receives the payload correctly.
func recorderReceivesPayload(t *testing.T, cons Constructor) {
	ctx := context.Background()
	cons.SetName(name)
	cons.SetIndexName(indexName)
	cons.SetTimeout(time.Second)
	cons.SetEndpoint(cons.TestServer().URL)
	rec, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if rec == nil {
		t.Fatal("rec = (nil); want (DataRecorder)")
	}
	err = rec.Ping()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	p := datatype.New([]datatype.DataType{})
	payload := recorder.Job{
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
		t.Error("didn't receive the payload")
	}
}

// recorderSendsResult tests the recorder send the results to the endpoint.
func recorderSendsResult(t *testing.T, cons Constructor) {
	ctx := context.Background()
	cons.SetName(name)
	cons.SetIndexName(indexName)
	cons.SetTimeout(time.Second)
	cons.SetEndpoint(cons.TestServer().URL)
	rec, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	err = rec.Ping()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	p := datatype.New([]datatype.DataType{&datatype.StringType{Key: "test", Value: "test"}})
	payload := recorder.Job{
		ID:        token.NewUID(),
		Payload:   p,
		IndexName: indexName,
		TypeName:  "my_type",
		Time:      time.Now(),
	}
	err = rec.Record(ctx, payload)
	if errors.Cause(err) != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
}

type badType struct{}

var errHappened = errors.New("this is bad")

func (badType) Read(b []byte) (int, error)         { return 0, errHappened }
func (badType) Equal(other datatype.DataType) bool { return false }
func (badType) Reset()                             {}

func errorsOnBadPayload(t *testing.T, cons Constructor) {
	ctx := context.Background()
	cons.SetName(name)
	cons.SetIndexName(indexName)
	cons.SetTimeout(time.Second)
	cons.SetEndpoint(cons.TestServer().URL)
	rec, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	err = rec.Ping()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	p := datatype.New([]datatype.DataType{&badType{}})
	payload := recorder.Job{
		ID:        token.NewUID(),
		Payload:   p,
		IndexName: indexName,
		TypeName:  "my_type",
		Time:      time.Now(),
	}
	err = rec.Record(ctx, payload)
	if errors.Cause(err) == nil {
		t.Errorf("err = (nil); want (%#v)", errHappened)
	}
}

func errorRecordingUnavailableEndpoint(t *testing.T, cons Constructor) {
	ctx := context.Background()
	ts := cons.TestServer()
	cons.SetName(name)
	cons.SetIndexName(indexName)
	cons.SetTimeout(time.Second)
	cons.SetEndpoint(ts.URL)
	rec, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	err = rec.Ping()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	p := datatype.New([]datatype.DataType{})
	payload := recorder.Job{
		ID:        token.NewUID(),
		Payload:   p,
		IndexName: indexName,
		TypeName:  "my_type",
		Time:      time.Now(),
	}
	ts.Close()
	err = rec.Record(ctx, payload)
	if errors.Cause(err) == nil {
		t.Errorf("err = (nil); want (%#v)", errHappened)
	}
}
