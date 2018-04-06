// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/arsham/expipe/recorder"
	"github.com/pkg/errors"
)

func TestInvalidEndpointError(t *testing.T) {
	msg := "the endpoint"
	e := recorder.InvalidEndpointError(msg)
	check(t, e.Error(), msg)
}

func TestEndpointNotAvailableError(t *testing.T) {
	endpoint := "the endpoint"
	err := errors.New("my error")
	e := recorder.EndpointNotAvailableError{Endpoint: endpoint, Err: err}
	check(t, e.Error(), endpoint)
	check(t, e.Error(), err.Error())
}

func TestParseTimeOutError(t *testing.T) {
	timeout := "5"
	err := errors.New("my error")
	e := recorder.ParseTimeOutError{Timeout: timeout, Err: err}
	check(t, e.Error(), timeout)
	check(t, e.Error(), err.Error())
}

func TestInvalidIndexNameError(t *testing.T) {
	indexName := "thumb is not an index finger"
	e := recorder.InvalidIndexNameError(indexName)
	check(t, e.Error(), indexName)
}

func TestLowTimeoutError(t *testing.T) {
	timeout := 5
	e := recorder.LowTimeout(timeout)
	check(t, e.Error(), strconv.Itoa(timeout))
}

func check(t *testing.T, err, msg string) {
	if !strings.Contains(err, msg) {
		t.Errorf("Contains(err, msg) = false: want (%s) to be in (%s)", msg, err)
	}
}
