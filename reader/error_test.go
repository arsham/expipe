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

func TestInvalidEndpointError(t *testing.T) {
	msg := "the endpoint"
	e := reader.InvalidEndpointError(msg)
	check(t, e.Error(), msg)
}

func TestEndpointNotAvailableError(t *testing.T) {
	endpoint := "the endpoint"
	err := errors.New("my error")
	e := reader.EndpointNotAvailableError{Endpoint: endpoint, Err: err}
	check(t, e.Error(), endpoint)
	check(t, e.Error(), err.Error())
}

func TestLowIntervalError(t *testing.T) {
	interval := 5
	e := reader.LowIntervalError(interval)
	check(t, e.Error(), strconv.Itoa(interval))
}

func TestLowTimeoutError(t *testing.T) {
	timeout := 5
	e := reader.LowTimeoutError(timeout)
	check(t, e.Error(), strconv.Itoa(timeout))
}

func check(t *testing.T, err, msg string) {
	if !strings.Contains(err, msg) {
		t.Errorf("Contains(err, msg): want (%s) to be in (%s)", msg, err)
	}
}
