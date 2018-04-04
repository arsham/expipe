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

// pingingEndpoint is a helper to test the reader errors when the endpoint
// goes away.
func pingingEndpoint(t testing.TB, cons Constructor) {
	unavailableEndpoint := "http://192.168.255.255"
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
		t.Errorf("err.(reader.EndpointNotAvailableError = (%#v); want (reader.EndpointNotAvailableError)", err)
	}
	cons.SetEndpoint(unavailableEndpoint)
	red, _ = cons.Object()
	err = red.Ping()
	if _, ok := errors.Cause(err).(reader.EndpointNotAvailableError); !ok {
		t.Error("err.(reader.EndpointNotAvailableError = (nil); want (reader.EndpointNotAvailableError)")
	}
	if !strings.Contains(err.Error(), unavailableEndpoint) {
		t.Errorf("Contains(err, unavailableEndpoint): want (%s) to be in (%s)", unavailableEndpoint, err.Error())
	}
}

// readerErrorsOnEndpointDisapears is a helper to test the reader errors
// when the endpoint goes away.
func readerErrorsOnEndpointDisapears(t testing.TB, cons Constructor) {
	var ok bool
	ctx := context.Background()
	ts := cons.TestServer()
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetEndpoint(ts.URL)
	cons.SetInterval(time.Hour)
	cons.SetTimeout(time.Hour)
	cons.SetBackoff(5)
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

// readerBacksOffOnEndpointGone is a helper to test the reader backs off
// when the endpoint goes away.
func readerBacksOffOnEndpointGone(t testing.TB, cons Constructor) {
	ts := cons.TestServer()
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetEndpoint(ts.URL)
	cons.SetInterval(time.Hour)
	cons.SetTimeout(time.Hour)
	cons.SetLogger(tools.DiscardLogger())
	cons.SetBackoff(5)
	red, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	err = red.Ping()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}

	ts.Close()
	ctx := context.Background()
	job := token.New(ctx)

	backedOff := false
	// We don't know the backoff amount set in the reader, so we try
	// 100 times until it closes.
	for i := 0; i < 100; i++ {
		_, err = red.Read(job)
		if err == reader.ErrBackoffExceeded {
			backedOff = true
			break
		}
	}
	r, err := red.Read(job)
	if errors.Cause(err) == nil {
		t.Error("err = (nil); want (error)")
	}
	if r != nil {
		t.Errorf("r = (%#v); want (nil)", r)
	}

	t.Skip("check this out")
	if !backedOff {
		t.Error("want (true), got (false)")
	}
}

// readingReturnsErrorIfNotPingedYet is a helper to test the reader
// returns an error if the caller hasn't called the Ping() method.
func readingReturnsErrorIfNotPingedYet(t testing.TB, cons Constructor) {
	ctx := context.Background()
	job := token.New(ctx)
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetInterval(time.Second)
	cons.SetTimeout(time.Second)
	cons.SetBackoff(5)

	red, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	res, err := red.Read(job)
	if errors.Cause(err) != reader.ErrPingNotCalled {
		t.Errorf("err = (%#v); want (reader.ErrPingNotCalled)", err)
	}
	if res != nil {
		t.Errorf("res = (%#v); want (nil)", res)
	}
}
