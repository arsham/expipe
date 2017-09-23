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

	"github.com/arsham/expipe/reader"
	"github.com/pkg/errors"
)

const (
	name     = "the name"
	typeName = "my type"
)

func testShouldNotChangeTheInput(t *testing.T, cons Constructor) {
	endpoint := cons.TestServer().URL
	interval := time.Hour
	timeout := time.Hour
	backoff := 5
	cons.SetName(name)
	cons.SetTypeName(typeName)
	cons.SetEndpoint(endpoint)
	cons.SetInterval(interval)
	cons.SetTimeout(timeout)
	cons.SetBackoff(backoff)

	red, err := cons.Object()
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}

	if red.Name() != name {
		t.Errorf("given name should not be changed: %v", red.Name())
	}
	if red.TypeName() != typeName {
		t.Errorf("given type name should not be changed: %v", red.TypeName())
	}
	if red.Interval() != interval {
		t.Errorf("given interval should not be changed: %v", red.Timeout())
	}
	if red.Timeout() != timeout {
		t.Errorf("given timeout should not be changed: %v", red.Timeout())
	}
}

func testNameCheck(t *testing.T, cons Constructor) {
	endpoint := cons.TestServer().URL
	cons.SetName("")
	cons.SetEndpoint(endpoint)

	red, err := cons.Object()
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
	}
	if errors.Cause(err) != reader.ErrEmptyName {
		t.Errorf("expected ErrEmptyName, got (%v)", err)
	}
}

func testTypeNameCheck(t *testing.T, cons Constructor) {
	endpoint := cons.TestServer().URL
	cons.SetName(name)
	cons.SetTypeName("")
	cons.SetEndpoint(endpoint)

	red, err := cons.Object()
	if !reflect.ValueOf(red).IsNil() {
		t.Fatalf("expected nil, got (%v)", red)
	}
	if errors.Cause(err) != reader.ErrEmptyTypeName {
		t.Errorf("expected ErrEmptyTypeName, got (%v)", err)
	}
}

func testBackoffCheck(t *testing.T, cons Constructor) {
	endpoint := cons.TestServer().URL
	backoff := 3
	cons.SetEndpoint(endpoint)
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetBackoff(backoff)

	red, err := cons.Object()
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%#v)", red)
	}
	err = errors.Cause(err)
	if _, ok := err.(reader.ErrLowBackoffValue); !ok {
		t.Fatalf("expected ErrLowBackoffValue, got (%v)", err)
	}
	if !strings.Contains(err.Error(), strconv.Itoa(backoff)) {
		t.Errorf("expected (%d) be mentioned, got (%v)", backoff, err)
	}
}

func testIntervalCheck(t *testing.T, cons Constructor) {
	endpoint := cons.TestServer().URL
	interval := 0
	cons.SetEndpoint(endpoint)
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetInterval(time.Duration(interval))

	red, err := cons.Object()
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
	}
	err = errors.Cause(err)
	if _, ok := err.(reader.ErrLowInterval); !ok {
		t.Fatalf("expected ErrLowInterval, got (%v)", err)
	}
	if !strings.Contains(err.Error(), strconv.Itoa(interval)) {
		t.Errorf("expected (%d) be mentioned, got (%v)", interval, err)
	}
}

func testEndpointCheck(t *testing.T, cons Constructor) {
	cons.SetTypeName("my type")
	cons.SetEndpoint("")

	red, err := cons.Object()
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
	}
	if errors.Cause(err) != reader.ErrEmptyEndpoint {
		t.Errorf("expected ErrEmptyEndpoint, got (%v)", err)
	}

	const invalidEndpoint = "this is invalid"
	cons.SetEndpoint(invalidEndpoint)
	red, err = cons.Object()
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
	}
	err = errors.Cause(err)
	if _, ok := err.(reader.ErrInvalidEndpoint); !ok {
		t.Fatalf("expected ErrInvalidEndpoint, got (%T)", err)
	}
	if !strings.Contains(err.Error(), invalidEndpoint) {
		t.Errorf("expected (%s) be in the error message, got (%v)", invalidEndpoint, err)
	}
}
