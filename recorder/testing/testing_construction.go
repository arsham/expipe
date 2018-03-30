// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/recorder"
	"github.com/pkg/errors"
)

func shouldNotChangeTheInput(t *testing.T, cons Constructor) {
	name := "recorder name"
	indexName := "recorder_index_name"
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
		t.Errorf("rec.Timeout() = (%s); want (%s)", rec.Timeout(), timeout)
	}
}

func backoffCheck(t *testing.T, cons Constructor) {
	backoff := 3
	cons.SetName("the name")
	cons.SetIndexName("my_index_name")
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetTimeout(time.Second)
	cons.SetBackoff(backoff)
	rec, err := cons.Object()
	if err == nil {
		t.Error("err = (nil); want (recorder.LowBackoffValueError)")
	}
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("rec = (%#v); want (nil)", rec)
	}
	if !strings.Contains(err.Error(), strconv.Itoa(backoff)) {
		t.Errorf("Contains(err, backoff): want (%d) to be in (%v)", backoff, err.Error())
	}
}

func nameCheck(t *testing.T, cons Constructor) {
	indexName := "the_index_name"
	cons.SetIndexName(indexName)
	cons.SetTimeout(time.Hour)
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetBackoff(5)
	cons.SetName("")
	rec, err := cons.Object()
	if errors.Cause(err) != recorder.ErrEmptyName {
		t.Errorf("err = (%#v); want (recorder.ErrEmptyName)", err)
	}
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("rec = (%#v); want (nil)", rec)
	}
}

func indexNameCheck(t *testing.T, cons Constructor) {
	name := "the name"
	indexName := "index_name"
	cons.SetName(name)
	cons.SetIndexName(indexName)
	cons.SetTimeout(time.Hour)
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetBackoff(5)
	cons.SetIndexName("aa bb")
	rec, err := cons.Object()
	if _, ok := errors.Cause(err).(recorder.InvalidIndexNameError); !ok {
		t.Errorf("err = (%#v); want (recorder.InvalidIndexNameError)", err)
	}
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("rec = (%#v); want (nil)", rec)
	}
}

func endpointCheck(t *testing.T, cons Constructor) {
	invalidEndpoint := "this is invalid"
	cons.SetName("the name")
	cons.SetIndexName("my_index_name")
	cons.SetTimeout(time.Hour)
	cons.SetBackoff(5)
	cons.SetEndpoint(invalidEndpoint)
	rec, err := cons.Object()
	if _, ok := errors.Cause(err).(recorder.InvalidEndpointError); !ok {
		t.Errorf("err = (%#v); want (recorder.InvalidEndpointError)", err)
	}
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("rec = (%#v); want (nil)", rec)
	}
	if !strings.Contains(err.Error(), invalidEndpoint) {
		t.Errorf("Contains(err, invalidEndpoint): want (%s) be in (%#v)", invalidEndpoint, err)
	}
}
