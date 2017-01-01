// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func testRecorderErrorsOnInvalidEndpoint(t *testing.T, setup setupFunc) {
	name := "the name"
	indexName := "index_name"
	backoff := 5

	timeout := 10 * time.Millisecond
	if isTravis() {
		timeout = 10 * time.Second
	}
	invalidEndpoint := "this is invalid"
	rec, err := setup(RecorderErrorsOnInvalidEndpointTestCase, name, invalidEndpoint, indexName, timeout, backoff)
	if rec == nil {
		t.Fatal("You should implement RecorderErrorsOnInvalidEndpointTestCase")
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

func testRecorderConstructionCases(t *testing.T, setup setupFunc) {
	name := "the name"
	indexName := "index_name"
	endpoint := "http://127.0.0.1:9200"
	backoff := 5

	timeout := 10 * time.Millisecond
	if isTravis() {
		timeout = 10 * time.Second
	}
	rec, _ := setup(RecorderConstructionCasesTestCase, name, endpoint, indexName, timeout, backoff)

	if rec.Name() != name {
		t.Errorf("given name should not be changed: %v", rec.Name())
	}
	if rec.IndexName() != indexName {
		t.Errorf("given index name should not be changed: %v", rec.IndexName())
	}
	if rec.Timeout() != timeout {
		t.Errorf("given timeout should not be changed: %v", rec.Timeout())
	}

	// Backoff check
	rec, err := setup(RecorderConstructionCasesTestCase, name, endpoint, indexName, timeout, 3)
	if !reflect.ValueOf(rec).IsNil() {
		t.Errorf("expected nil, got (%#v)", rec)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(interface {
		LowBackoffValue()
	}); !ok {
		t.Errorf("expected ErrLowBackoffValue, got (%v)", err)
	}
	if !strings.Contains(err.Error(), "3") {
		t.Errorf("expected 3 be mentioned, got (%v)", err)
	}
}
