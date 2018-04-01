// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder_test

import (
	"testing"
	"time"

	"github.com/arsham/expipe/recorder"
	recorder_testing "github.com/arsham/expipe/recorder/testing"
	"github.com/arsham/expipe/tools"
	"github.com/pkg/errors"
)

func TestSetLogger(t *testing.T) {
	r := recorder_testing.Recorder{}
	err := recorder.WithLogger(nil)(&r)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}

	err = recorder.WithLogger(tools.DiscardLogger())(&r)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
}

func TestSetName(t *testing.T) {
	r := recorder_testing.Recorder{}
	err := recorder.WithName("")(&r)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}

	err = recorder.WithName("name")(&r)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
}

func TestSetEndpoint(t *testing.T) {
	r := recorder_testing.Recorder{}
	err := recorder.WithEndpoint("")(&r)
	err = errors.Cause(err)
	if err != recorder.ErrEmptyEndpoint {
		t.Errorf("err = (%T); want (recorder.ErrEmptyEndpoint)", err)
	}

	err = recorder.WithEndpoint("invalid endpoint")(&r)
	err = errors.Cause(err)
	if _, ok := err.(recorder.InvalidEndpointError); !ok {
		t.Errorf("err = (%T); want (recorder.InvalidEndpointError)", err)
	}

	err = recorder.WithEndpoint("http://localhost")(&r)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
}

func TestSetIndexName(t *testing.T) {
	r := recorder_testing.Recorder{}
	err := recorder.WithIndexName("")(&r)
	if errors.Cause(err) != recorder.ErrEmptyIndexName {
		t.Errorf("err = (%v); want (recorder.ErrEmptyIndexName)", err)
	}

	err = recorder.WithIndexName("a b")(&r)
	if _, ok := errors.Cause(err).(recorder.InvalidIndexNameError); !ok {
		t.Errorf("err = (%v); want (recorder.InvalidIndexNameError)", err)
	}

	err = recorder.WithIndexName("name")(&r)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
}

func TestSetTimeout(t *testing.T) {
	r := recorder_testing.Recorder{}
	err := recorder.WithTimeout(time.Duration(0))(&r)
	if _, ok := errors.Cause(err).(recorder.LowTimeout); !ok {
		t.Errorf("err = (%v); want (recorder.LowTimeoutError)", err)
	}

	err = recorder.WithTimeout(time.Millisecond * 10)(&r)
	if _, ok := errors.Cause(err).(recorder.LowTimeout); !ok {
		t.Errorf("err = (%v); want (recorder.LowTimeoutError)", err)
	}

	err = recorder.WithTimeout(time.Second)(&r)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
}

func TestSetBackoff(t *testing.T) {
	r := recorder_testing.Recorder{}
	err := recorder.WithBackoff(4)(&r)
	if _, ok := errors.Cause(err).(recorder.LowBackoffValueError); !ok {
		t.Errorf("err = (%v); want (recorder.LowBackoffValueError)", err)
	}

	err = recorder.WithBackoff(5)(&r)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
}
