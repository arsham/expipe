// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder_test

import (
	"testing"
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/recorder"
	recorder_testing "github.com/arsham/expipe/recorder/testing"
	"github.com/pkg/errors"
)

func TestSetLogger(t *testing.T) {
	r := recorder_testing.Recorder{}
	err := recorder.SetLogger(nil)(&r)
	if err == nil {
		t.Error("want (error), got (nil)")
	}

	err = recorder.SetLogger(internal.DiscardLogger())(&r)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}

func TestSetName(t *testing.T) {
	r := recorder_testing.Recorder{}
	err := recorder.SetName("")(&r)
	if err == nil {
		t.Error("want (error), got (nil)")
	}

	err = recorder.SetName("name")(&r)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}

func TestSetEndpoint(t *testing.T) {
	r := recorder_testing.Recorder{}
	err := recorder.SetEndpoint("")(&r)
	err = errors.Cause(err)
	if err != recorder.ErrEmptyEndpoint {
		t.Errorf("want (recorder.ErrEmptyEndpoint), got (%T)", err)
	}

	err = recorder.SetEndpoint("invalid endpoint")(&r)
	err = errors.Cause(err)
	if _, ok := err.(recorder.ErrInvalidEndpoint); !ok {
		t.Errorf("want (recorder.ErrInvalidEndpoint), got (%T)", err)
	}

	err = recorder.SetEndpoint("http://localhost")(&r)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}

func TestSetIndexName(t *testing.T) {
	r := recorder_testing.Recorder{}
	err := recorder.SetIndexName("")(&r)
	if errors.Cause(err) != recorder.ErrEmptyIndexName {
		t.Errorf("want (recorder.ErrEmptyIndexName), got (%v)", err)
	}

	err = recorder.SetIndexName("a b")(&r)
	if _, ok := errors.Cause(err).(recorder.ErrInvalidIndexName); !ok {
		t.Errorf("want (recorder.ErrInvalidIndexName), got (%v)", err)
	}

	err = recorder.SetIndexName("name")(&r)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}

func TestSetTimeout(t *testing.T) {
	r := recorder_testing.Recorder{}
	err := recorder.SetTimeout(time.Duration(0))(&r)
	if _, ok := errors.Cause(err).(recorder.ErrLowTimeout); !ok {
		t.Errorf("want (recorder.ErrLowTimeout), got (%v)", err)
	}

	err = recorder.SetTimeout(time.Millisecond * 10)(&r)
	if _, ok := errors.Cause(err).(recorder.ErrLowTimeout); !ok {
		t.Errorf("want (recorder.ErrLowTimeout), got (%v)", err)
	}

	err = recorder.SetTimeout(time.Second)(&r)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}

func TestSetBackoff(t *testing.T) {
	r := recorder_testing.Recorder{}
	err := recorder.SetBackoff(4)(&r)
	if _, ok := errors.Cause(err).(recorder.ErrLowBackoffValue); !ok {
		t.Errorf("want (recorder.ErrLowBackoffValue), got (%v)", err)
	}

	err = recorder.SetBackoff(5)(&r)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}
