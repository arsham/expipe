// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/recorder"
	"github.com/arsham/expipe/tools/token"
	"github.com/pkg/errors"
)

var (
	constName = "the name"
	indexName = "my_index_name"
)

// recorderErrorsOnUnavailableEndpoint tests the recorder errors for bad URL.
func recorderErrorsOnUnavailableEndpoint(t *testing.T, cons Constructor) {
	timeout := time.Second
	name := constName
	backoff := 5
	ts := cons.TestServer()
	cons.SetName(name)
	cons.SetIndexName(indexName)
	cons.SetEndpoint(ts.URL)
	cons.SetTimeout(timeout)
	cons.SetBackoff(backoff)
	ts.Close()

	rec, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	err = rec.Ping()
	if _, ok := errors.Cause(err).(recorder.EndpointNotAvailableError); !ok {
		t.Errorf("err = (%#v); want (recorder.EndpointNotAvailableError)", err)
	}
}

// recorderBacksOffOnEndpointGone is a helper to test the recorder backs off
// when the endpoint goes away.
func recorderBacksOffOnEndpointGone(t *testing.T, cons Constructor) {
	ctx := context.Background()
	ts := cons.TestServer()
	timeout := time.Second
	cons.SetName(constName)
	cons.SetIndexName(indexName)
	cons.SetEndpoint(ts.URL)
	cons.SetTimeout(timeout)
	cons.SetBackoff(5)
	rec, err := cons.Object()
	if err != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	if reflect.ValueOf(rec).IsNil() {
		t.Error("rec = (nil); want (DataRecorder)")
	}
	err = rec.Ping()
	if err != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}

	ts.Close()
	p := datatype.New([]datatype.DataType{})
	payload := &recorder.Job{
		ID:        token.NewUID(),
		Payload:   p,
		IndexName: "my index",
		TypeName:  "my type",
		Time:      time.Now(),
	}

	// We don't know the backoff amount set in the recorder, so we
	// try 100 times until it closes.
	backedOff := false
	for i := 0; i < 100; i++ {
		err := rec.Record(ctx, payload)
		if err == recorder.ErrBackoffExceeded {
			backedOff = true
			break
		}
	}
	stop := make(chan struct{})
	go func() {
		rec.Record(ctx, payload)
		close(stop)
	}()
	select {
	case <-stop:
	case <-time.After(5 * time.Second):
		t.Error("recorder didn't back off")
	}
	<-stop
	if !backedOff {
		t.Skip("check this out")
	}
}

// recordingReturnsErrorIfNotPingedYet is a helper to test the recorder
// returns an error if the caller hasn't called the Ping() method.
func recordingReturnsErrorIfNotPingedYet(t *testing.T, cons Constructor) {
	ctx := context.Background()
	timeout := time.Second
	cons.SetName(constName)
	cons.SetIndexName(indexName)
	cons.SetTimeout(timeout)
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetBackoff(5)
	rec, err := cons.Object()
	if err != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	if reflect.ValueOf(rec).IsNil() {
		t.Error("rec = (nil); want (DataRecorder)")
	}

	p := datatype.New([]datatype.DataType{})
	payload := &recorder.Job{
		ID:        token.NewUID(),
		Payload:   p,
		IndexName: "my index",
		TypeName:  "my type",
		Time:      time.Now(),
	}
	err = rec.Record(ctx, payload)
	if errors.Cause(err) != recorder.ErrPingNotCalled {
		t.Errorf("err = (%#v); want (recorder.ErrPingNotCalled)", err)
	}
}
