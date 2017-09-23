// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/arsham/expipe/recorder"
)

func TestErrInvalidEndpoint(t *testing.T) {
	msg := "the endpoint"
	e := recorder.ErrInvalidEndpoint(msg)
	if !strings.Contains(e.Error(), msg) {
		t.Errorf("want (%s) in error, got (%s)", msg, e.Error())
	}
}

func TestErrEndpointNotAvailable(t *testing.T) {
	endpoint := "the endpoint"
	err := errors.New("my error")
	e := recorder.ErrEndpointNotAvailable{Endpoint: endpoint, Err: err}
	if !strings.Contains(e.Error(), endpoint) {
		t.Errorf("want (%s) in error, got (%s)", endpoint, e.Error())
	}
	if !strings.Contains(e.Error(), err.Error()) {
		t.Errorf("want (%s) in error, got (%s)", err.Error(), e.Error())
	}
}

func TestErrLowBackoffValue(t *testing.T) {
	backoff := 5
	e := recorder.ErrLowBackoffValue(backoff)
	if !strings.Contains(e.Error(), strconv.Itoa(backoff)) {
		t.Errorf("want (%s) in error, got (%s)", strconv.Itoa(backoff), e.Error())
	}
}

func TestErrParseInterval(t *testing.T) {
	interval := "5"
	err := errors.New("my error")
	e := recorder.ErrParseInterval{Interval: interval, Err: err}
	if !strings.Contains(e.Error(), interval) {
		t.Errorf("want (%s) in error, got (%s)", interval, e.Error())
	}
	if !strings.Contains(e.Error(), err.Error()) {
		t.Errorf("want (%s) in error, got (%s)", err.Error(), e.Error())
	}
}

func TestErrParseTimeOut(t *testing.T) {
	timeout := "5"
	err := errors.New("my error")
	e := recorder.ErrParseTimeOut{Timeout: timeout, Err: err}
	if !strings.Contains(e.Error(), timeout) {
		t.Errorf("want (%s) in error, got (%s)", timeout, e.Error())
	}
	if !strings.Contains(e.Error(), err.Error()) {
		t.Errorf("want (%s) in error, got (%s)", err.Error(), e.Error())
	}
}

func TestErrInvalidIndexName(t *testing.T) {
	indexName := "thumb is not an index finger"
	e := recorder.ErrInvalidIndexName(indexName)
	if !strings.Contains(e.Error(), indexName) {
		t.Errorf("want (%s) in error, got (%s)", indexName, e.Error())
	}
}

func TestErrLowTimeout(t *testing.T) {
	timeout := 5
	e := recorder.ErrLowTimeout(timeout)
	if !strings.Contains(e.Error(), strconv.Itoa(timeout)) {
		t.Errorf("want (%s) in error, got (%s)", strconv.Itoa(timeout), e.Error())
	}
}
