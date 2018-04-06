// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/recorder"
	"github.com/arsham/expipe/tools"
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
	ts := cons.TestServer()
	cons.SetName(name)
	cons.SetIndexName(indexName)
	cons.SetEndpoint(ts.URL)
	cons.SetTimeout(timeout)
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

// recordingReturnsErrorIfNotPingedYet is a helper to test the recorder returns
// an error if the caller hasn't called the Ping() method.
func recordingReturnsErrorIfNotPingedYet(t *testing.T, cons Constructor) {
	ctx := context.Background()
	timeout := time.Second
	cons.SetName(constName)
	cons.SetIndexName(indexName)
	cons.SetTimeout(timeout)
	cons.SetEndpoint(cons.TestServer().URL)
	rec, err := cons.Object()
	if err != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	if reflect.ValueOf(rec).IsNil() {
		t.Error("rec = (nil); want (DataRecorder)")
	}

	p := datatype.New([]datatype.DataType{})
	payload := recorder.Job{
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

	err = rec.Ping()
	if err != nil {
		t.Fatalf("Ping() = (%v); want (nil)", err)
	}
	err = rec.Record(ctx, payload)
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
}

func pingingEndpoint(t testing.TB, cons Constructor) {
	if testing.Short() {
		return
	}
	ts := cons.TestServer()
	cons.SetName(name)
	cons.SetIndexName(indexName)
	cons.SetEndpoint(ts.URL)
	cons.SetTimeout(time.Second)
	cons.SetLogger(tools.DiscardLogger())
	ts.Close()

	rec, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	if rec == nil {
		t.Fatal("rec = (nil); want (DataRecorder)")
	}
	err = rec.Ping()
	if _, ok := errors.Cause(err).(recorder.EndpointNotAvailableError); !ok {
		t.Errorf("err = (%#v); want (recorder.EndpointNotAvailableError)", err)
	}
	cons.SetEndpoint(ts.URL)
	rec, _ = cons.Object()
	err = rec.Ping()
	if _, ok := errors.Cause(err).(recorder.EndpointNotAvailableError); !ok {
		t.Error("err = (nil); want (recorder.EndpointNotAvailableError)")
	}
	if !strings.Contains(err.Error(), ts.URL) {
		t.Errorf("Contains(): want (%s) to be in (%s)", ts.URL, err.Error())
	}
}
