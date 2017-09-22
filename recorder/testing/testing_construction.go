// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/arsham/expipe/recorder"
	"github.com/pkg/errors"
)

func testShowNotChangeTheInput(t *testing.T, cons Constructor) {
	name := "the name"
	indexName := "my_index_name"
	endpoint := cons.TestServer().URL
	timeout := 10 * time.Millisecond
	backoff := 5
	cons.SetName(name)
	cons.SetIndexName(indexName)
	cons.SetEndpoint(endpoint)
	cons.SetTimeout(timeout)
	cons.SetBackoff(backoff)

	rec, err := cons.Object()
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}

	if rec.Name() != name {
		t.Errorf("given name should not be changed: %v", rec.Name())
	}
	if rec.IndexName() != indexName {
		t.Errorf("given index name should not be changed: %v", rec.IndexName())
	}
	if rec.Timeout() != timeout {
		t.Errorf("given timeout should not be changed: %v", rec.Timeout())
	}
}

func testBackoffCheck(t *testing.T, cons Constructor) {
	cons.SetName("the name")
	cons.SetIndexName("my_index_name")
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetTimeout(10 * time.Millisecond)

	cons.SetBackoff(3)
	rec, err := cons.Object()
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("expected nil, got (%#v)", rec)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	err = errors.Cause(err)
	if _, ok := err.(interface {
		LowBackoffValue()
	}); !ok {
		t.Errorf("expected ErrLowBackoffValue, got (%v)", err)
	}
	if !strings.Contains(err.Error(), "3") {
		t.Errorf("expected 3 be mentioned, got (%v)", err)
	}
}

func testNameCheck(t *testing.T, cons Constructor) {
	name := "the name"
	indexName := "index_name"
	cons.SetName("")
	cons.SetIndexName(indexName)
	cons.SetTimeout(time.Hour)
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetBackoff(5)

	rec, err := cons.Object()
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("expected nil, got (%#v)", rec)
	}
	if err != recorder.ErrEmptyName {
		t.Errorf("expected ErrEmptyName, got (%v)", err)
	}

	cons.SetName(name)
	cons.SetIndexName("")
	rec, err = cons.Object()
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("expected nil, got (%#v)", rec)
	}
	if err != recorder.ErrEmptyIndexName {
		t.Errorf("expected ErrEmptyIndexName, got (%v)", err)
	}

	cons.SetName(name)
	cons.SetIndexName("aa bb")
	rec, err = cons.Object()
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("expected nil, got (%#v)", rec)
	}
	if _, ok := err.(interface {
		InvalidIndexName()
	}); !ok {
		t.Errorf("expected InvalidIndexName, got (%v)", err)
	}
}

func testEndpointCheck(t *testing.T, cons Constructor) {

	invalidEndpoint := "this is invalid"
	cons.SetName("the name")
	cons.SetIndexName("my_index_name")
	cons.SetTimeout(time.Hour)
	cons.SetBackoff(5)
	cons.SetEndpoint(invalidEndpoint)

	rec, err := cons.Object()
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("expected nil, got (%#v)", rec)
	}
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("expected nil, got (%v)", rec)
	}
	if _, ok := err.(interface {
		InvalidEndpoint()
	}); !ok {
		t.Fatalf("expected ErrInvalidEndpoint, got (%v)", err)
	}
	if !strings.Contains(err.Error(), invalidEndpoint) {
		t.Errorf("expected (%s) be in the error message, got (%v)", invalidEndpoint, err)
	}
}
