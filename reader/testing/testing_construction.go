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

func testShowNotChangeTheInput(t *testing.T, setup setupFunc, name string, typeName string, endpoint string, interval time.Duration, timeout time.Duration, backoff int) {
	red, err := setup(name, endpoint, typeName, interval, timeout, backoff)
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

func testNameCheck(t *testing.T, setup setupFunc, name string, typeName string, endpoint string, interval time.Duration, timeout time.Duration, backoff int) {
	red, err := setup("", endpoint, typeName, interval, timeout, backoff)
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%#v)", red)
	}
	if err != reader.ErrEmptyName {
		t.Errorf("expected ErrEmptyName, got (%v)", err)
	}

	// TypeName Check
	red, err = setup(name, endpoint, "", interval, timeout, backoff)
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%#v)", red)
	}
	if err != reader.ErrEmptyTypeName {
		t.Errorf("expected ErrEmptyTypeName, got (%v)", err)
	}
}

func testBackoffCheck(t *testing.T, setup setupFunc, name string, typeName string, endpoint string, interval time.Duration, timeout time.Duration, backoff int) {
	red, err := setup(name, endpoint, typeName, interval, timeout, 3)
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

func testEndpointCheck(t *testing.T, setup setupFunc, name string, typeName string, endpoint string, interval time.Duration, timeout time.Duration, backoff int) {

	red, err := setup(name, "", typeName, interval, timeout, backoff)
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
	}
	if err != reader.ErrEmptyEndpoint {
		t.Errorf("expected ErrEmptyEndpoint, got (%v)", err)
	}

	const invalidEndpoint = "this is invalid"
	red, err = setup(name, invalidEndpoint, typeName, interval, timeout, backoff)
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

	unavailableEndpoint := "http://nowhere.localhost.localhost"
	red, err = setup(name, unavailableEndpoint, typeName, interval, timeout, backoff)
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("expected nil, got (%v)", red)
	}
	if err == nil {
		t.Fatal("expected ErrEndpointNotAvailable, got nil")
	}
	if _, ok := err.(interface {
		EndpointNotAvailable()
	}); !ok {
		t.Errorf("expected ErrEndpointNotAvailable, got (%v)", err)
	}
	if !strings.Contains(err.Error(), unavailableEndpoint) {
		t.Errorf("expected (%s) be in the error message, got (%v)", unavailableEndpoint, err)
	}
}
