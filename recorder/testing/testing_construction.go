// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/recorder"
	"github.com/pkg/errors"
)

const (
	name = "the name"
)

func shouldNotChangeTheInput(t *testing.T, cons Constructor) {
	endpoint := cons.TestServer().URL
	timeout := time.Second
	backoff := 5
	logger := internal.DiscardLogger()
	cons.SetName(name)
	cons.SetIndexName(indexName)
	cons.SetEndpoint(endpoint)
	cons.SetTimeout(timeout)
	cons.SetBackoff(backoff)
	cons.SetLogger(logger)
	rec, err := cons.Object()
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if rec.Name() != name {
		t.Errorf("rec.Name() = (%s); want (%s)", rec.Name(), name)
	}
	if rec.IndexName() != indexName {
		t.Errorf("rec.IndexName() = (%s); want (%s)", rec.IndexName(), indexName)
	}
	if rec.Timeout() != timeout {
		t.Errorf("rec.Timeout() = (%d); want (%d)", rec.Timeout(), timeout)
	}
	if rec.Endpoint() != endpoint {
		t.Errorf("rec.Endpoint() = (%s); want (%s)", rec.Endpoint(), endpoint)
	}
	if rec.Backoff() != backoff {
		t.Errorf("rec.Backoff() = (%d); want (%d)", rec.Backoff(), backoff)
	}
}

func nameCheck(t *testing.T, cons Constructor) {
	cons.SetIndexName(indexName)
	cons.SetTimeout(time.Hour)
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetBackoff(5)
	rec, err := cons.Object()
	if errors.Cause(err) != recorder.ErrEmptyName {
		t.Errorf("err = (%#v); want (recorder.ErrEmptyName)", err)
	}
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("rec = (%#v); want (nil)", rec)
	}

	cons.SetName("")
	rec, err = cons.Object()
	if errors.Cause(err) != recorder.ErrEmptyName {
		t.Errorf("err = (%#v); want (recorder.ErrEmptyName)", err)
	}
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("rec = (%#v); want (nil)", rec)
	}
}

func indexNameCheck(t *testing.T, cons Constructor) {
	cons.SetName(name)
	cons.SetTimeout(time.Hour)
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetBackoff(5)
	rec, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	if reflect.ValueOf(rec).IsNil() {
		t.Error("rec = (nil); want (DataRecorder)")
	} else if rec.IndexName() != rec.Name() {
		t.Errorf("IndexName() = (%s); want (%s)", rec.IndexName(), rec.Name())
	}

	tcs := []string{"*", "\\", "<", "|", ",", ">", "/", "?", `"`, ` `}
	for _, tc := range tcs {
		newIndex := fmt.Sprintf("before%safter", tc)
		cons.SetIndexName(newIndex)
		rec, err = cons.Object()
		if _, ok := errors.Cause(err).(recorder.InvalidIndexNameError); !ok {
			t.Errorf("err = (%#v); want (recorder.InvalidIndexNameError)", err)
		}
		if !reflect.ValueOf(rec).IsNil() {
			t.Errorf("rec = (%#v); want (nil)", rec)
		}
	}

	cons.SetIndexName(indexName)
	rec, err = cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	if reflect.ValueOf(rec).IsNil() {
		t.Error("rec = (nil); want (DataRecorder)")
	} else if rec.IndexName() != indexName {
		t.Errorf("IndexName() = (%s); want (%s)", rec.IndexName(), indexName)
	}
}

func backoffCheck(t *testing.T, cons Constructor) {
	backoff := 3
	cons.SetName(name)
	cons.SetIndexName(indexName)
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetTimeout(time.Second)
	rec, err := cons.Object()
	if reflect.ValueOf(rec).IsNil() {
		t.Error("rec = (nil); want (DataRecorder)")
	}
	if rec.Backoff() < 5 {
		t.Errorf("Backoff() = (%d); want (>=5)", rec.Backoff())
	}

	cons.SetBackoff(backoff)
	rec, err = cons.Object()
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("rec = (%v); want (nil)", rec)
	}
	if _, ok := errors.Cause(err).(recorder.LowBackoffValueError); !ok {
		t.Errorf("err.(recorder.LowBackoffValueError) = (%#v); want (recorder.LowBackoffValueError)", err)
	}
	if !strings.Contains(err.Error(), strconv.Itoa(backoff)) {
		t.Errorf("Contains(err.Error(), backoff): want (%s) to be in (%s)", strconv.Itoa(backoff), err.Error())
	}
}

func endpointCheck(t testing.TB, cons Constructor) {
	var ok bool
	invalidEndpoint := "this is invalid"
	cons.SetName(name)
	cons.SetIndexName(indexName)
	cons.SetTimeout(time.Second)
	cons.SetBackoff(5)

	rec, err := cons.Object()
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("rec = (%#v); want (nil)", rec)
	}
	if errors.Cause(err) != recorder.ErrEmptyEndpoint {
		t.Errorf("err = (%#v); want (recorder.ErrEmptyEndpoint)", err)
	}

	cons.SetEndpoint("")
	rec, err = cons.Object()
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("rec = (%v); want (nil)", rec)
	}
	if errors.Cause(err) != recorder.ErrEmptyEndpoint {
		t.Errorf("err = (%#v); want (recorder.ErrEmptyEndpoint)", err)
	}

	cons.SetEndpoint(invalidEndpoint)
	rec, err = cons.Object()
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("rec = (%v); want (nil)", rec)
	}
	if _, ok = errors.Cause(err).(recorder.InvalidEndpointError); !ok {
		t.Errorf("err.(recorder.InvalidEndpointError) = (%#v); want (recorder.InvalidEndpointError)", err)
	}
	if ok && !strings.Contains(err.Error(), invalidEndpoint) {
		t.Errorf("want (%s) to be in (%s)", invalidEndpoint, err.Error())
	}
}

func timeoutCheck(t testing.TB, cons Constructor) {
	cons.SetName(name)
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetBackoff(5)
	rec, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	if reflect.ValueOf(rec).IsNil() {
		t.Fatal("rec = (nil); want (DataReader)")
	}
	if rec.Timeout() != 5*time.Second {
		t.Errorf("Timeout() = (%s); want (%s)", rec.Timeout(), 5*time.Second)
	}
}
