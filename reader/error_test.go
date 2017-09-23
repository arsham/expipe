// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/arsham/expipe/reader"
	"github.com/pkg/errors"
)

func TestErrInvalidEndpoint(t *testing.T) {
	msg := "the endpoint"
	e := reader.ErrInvalidEndpoint(msg)
	if !strings.Contains(e.Error(), msg) {
		t.Errorf("want (%s) in error, got (%s)", msg, e.Error())
	}
}

func TestErrEndpointNotAvailable(t *testing.T) {
	endpoint := "the endpoint"
	err := errors.New("my error")
	e := reader.ErrEndpointNotAvailable{Endpoint: endpoint, Err: err}
	if !strings.Contains(e.Error(), endpoint) {
		t.Errorf("want (%s) in error, got (%s)", endpoint, e.Error())
	}
	if !strings.Contains(e.Error(), err.Error()) {
		t.Errorf("want (%s) in error, got (%s)", err.Error(), e.Error())
	}
}

func TestErrLowBackoffValue(t *testing.T) {
	backoff := 5
	e := reader.ErrLowBackoffValue(backoff)
	if !strings.Contains(e.Error(), strconv.Itoa(backoff)) {
		t.Errorf("want (%s) in error, got (%s)", strconv.Itoa(backoff), e.Error())
	}
}

func TestErrLowInterval(t *testing.T) {
	interval := 5
	e := reader.ErrLowInterval(interval)
	if !strings.Contains(e.Error(), strconv.Itoa(interval)) {
		t.Errorf("want (%s) in error, got (%s)", strconv.Itoa(interval), e.Error())
	}
}

func TestErrLowTimeout(t *testing.T) {
	timeout := 5
	e := reader.ErrLowTimeout(timeout)
	if !strings.Contains(e.Error(), strconv.Itoa(timeout)) {
		t.Errorf("want (%s) in error, got (%s)", strconv.Itoa(timeout), e.Error())
	}
}
