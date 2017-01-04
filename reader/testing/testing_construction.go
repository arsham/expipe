// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/arsham/expvastic/reader"
)

func testShowNotChangeTheInput(t *testing.T, cons Constructor) {
	name := "the name"
	typeName := "my type"
	endpoint := cons.TestServer().URL
	interval := time.Hour
	timeout := time.Hour
	backoff := 5
	cons.SetName(name)
	cons.SetTypename(typeName)
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
	name := "the name"
	typeName := "my type"
	cons.SetName("")
	cons.SetTypename(typeName)
	cons.SetInterval(time.Hour)
	cons.SetTimeout(time.Hour)
	cons.SetBackoff(5)

	red, err := cons.Object()
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%#v)", red)
	}
	if err != reader.ErrEmptyName {
		t.Errorf("expected ErrEmptyName, got (%v)", err)
	}

	cons.SetName(name)
	cons.SetTypename("")
	// TypeName Check
	red, err = cons.Object()
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%#v)", red)
	}
	if err != reader.ErrEmptyTypeName {
		t.Errorf("expected ErrEmptyTypeName, got (%v)", err)
	}
}

func testBackoffCheck(t *testing.T, cons Constructor) {
	cons.SetName("the name")
	cons.SetTypename("my type")
	cons.SetInterval(time.Hour)
	cons.SetTimeout(time.Hour)
	cons.SetBackoff(3)

	red, err := cons.Object()
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%#v)", red)
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

func testEndpointCheck(t *testing.T, cons Constructor) {
	cons.SetName("the name")
	cons.SetTypename("my type")
	cons.SetInterval(time.Hour)
	cons.SetTimeout(time.Hour)
	cons.SetBackoff(5)
	cons.SetEndpoint("")

	red, err := cons.Object()
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
	}
	if err != reader.ErrEmptyEndpoint {
		t.Errorf("expected ErrEmptyEndpoint, got (%v)", err)
	}

	const invalidEndpoint = "this is invalid"
	cons.SetEndpoint(invalidEndpoint)
	red, err = cons.Object()
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
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
