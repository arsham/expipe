// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/tools"
	"github.com/arsham/expipe/tools/token"
	"github.com/pkg/errors"
)

// Reader should return an error when the endpoint is not reachable when
// pinging.
func pingingEndpoint(t testing.TB, cons Constructor) {
	ts := cons.TestServer()
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetEndpoint(ts.URL)
	cons.SetInterval(time.Millisecond)
	cons.SetTimeout(time.Second)
	cons.SetLogger(tools.DiscardLogger())
	ts.Close()

	red, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	if red == nil {
		t.Fatal("red = (nil); want (DataReader)")
	}
	err = red.Ping()
	if _, ok := errors.Cause(err).(reader.EndpointNotAvailableError); !ok {
		t.Errorf("err = (%#v); want (reader.EndpointNotAvailableError)", err)
	}
	cons.SetEndpoint(ts.URL)
	red, _ = cons.Object()
	err = red.Ping()
	if _, ok := errors.Cause(err).(reader.EndpointNotAvailableError); !ok {
		t.Error("err = (nil); want (reader.EndpointNotAvailableError)")
	}
	if !strings.Contains(err.Error(), ts.URL) {
		t.Errorf("Contains(): want (%s) to be in (%s)", ts.URL, err.Error())
	}
}

// Testing the reader errors when the endpoint goes away.
func readerErrorsOnEndpointDisapears(t testing.TB, cons Constructor) {
	ctx := context.Background()
	ts := cons.TestServer()
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetEndpoint(ts.URL)
	cons.SetInterval(time.Hour)
	cons.SetTimeout(time.Hour)
	red, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	err = red.Ping()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	ts.Close()
	result, err := red.Read(token.New(ctx))
	var ok bool
	if _, ok = errors.Cause(err).(reader.EndpointNotAvailableError); !ok {
		t.Errorf("err.(reader.EndpointNotAvailableError) = (%#v); want (reader.EndpointNotAvailableError)", err)
	}
	if ok && !strings.Contains(err.Error(), ts.URL) {
		t.Errorf("Contains(err, ts.URL): want (%s) to be in (%s)", ts.URL, err.Error())
	}
	if result != nil {
		t.Errorf("result = (%#v); want (nil)", result)
	}
}

// Testing the reader returns an error if the caller hasn't called the Ping()
// method.
func readingReturnsErrorIfNotPingedYet(t testing.TB, cons Constructor) {
	ctx := context.Background()
	job := token.New(ctx)
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetInterval(time.Second)
	cons.SetTimeout(time.Second)

	red, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	res, err := red.Read(job)
	if errors.Cause(err) != reader.ErrPingNotCalled {
		t.Errorf("err = (%#v); want (reader.ErrPingNotCalled)", err)
	}
	if res != nil {
		t.Errorf("res = (%s); want (nil)", res)
	}
	red.Ping()
	res, err = red.Read(job)
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	if res == nil {
		t.Error("res = (nil); want (result)")
	}
}
