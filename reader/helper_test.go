// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader_test

import (
	"testing"
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/internal/datatype"
	"github.com/arsham/expipe/reader"
	reader_testing "github.com/arsham/expipe/reader/testing"
	"github.com/pkg/errors"
)

func TestSetLogger(t *testing.T) {
	r := reader_testing.Reader{}
	err := reader.SetLogger(nil)(&r)
	if err == nil {
		t.Error("want (error), got (nil)")
	}

	err = reader.SetLogger(internal.DiscardLogger())(&r)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}

func TestSetName(t *testing.T) {
	r := reader_testing.Reader{}
	err := reader.SetName("")(&r)
	if err == nil {
		t.Error("want (error), got (nil)")
	}

	err = reader.SetName("name")(&r)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}

func TestSetEndpoint(t *testing.T) {
	r := reader_testing.Reader{}
	err := reader.SetEndpoint("")(&r)
	err = errors.Cause(err)
	if err != reader.ErrEmptyEndpoint {
		t.Errorf("want (reader.ErrEmptyEndpoint), got (%T)", err)
	}

	err = reader.SetEndpoint("invalid endpoint")(&r)
	err = errors.Cause(err)
	if _, ok := err.(reader.ErrInvalidEndpoint); !ok {
		t.Errorf("want (reader.ErrInvalidEndpoint), got (%T)", err)
	}

	err = reader.SetEndpoint("http://localhost")(&r)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}

func TestSetMapper(t *testing.T) {
	r := reader_testing.Reader{}
	err := reader.SetMapper(nil)(&r)
	if err == nil {
		t.Error("want (error), got (nil)")
	}

	err = reader.SetMapper(&datatype.MapConvertMock{})(&r)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}

func TestSetTypeName(t *testing.T) {
	r := reader_testing.Reader{}
	err := reader.SetTypeName("")(&r)
	if errors.Cause(err) != reader.ErrEmptyTypeName {
		t.Errorf("want (reader.ErrEmptyTypeName), got (%v)", err)
	}

	err = reader.SetTypeName("name")(&r)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}

func TestSetInterval(t *testing.T) {
	r := reader_testing.Reader{}
	err := reader.SetInterval(time.Duration(0))(&r)
	if _, ok := errors.Cause(err).(reader.ErrLowInterval); !ok {
		t.Errorf("want (reader.ErrLowInterval), got (%v)", err)
	}

	err = reader.SetInterval(time.Second)(&r)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}

func TestSetTimeout(t *testing.T) {
	r := reader_testing.Reader{}
	err := reader.SetTimeout(time.Duration(0))(&r)
	if _, ok := errors.Cause(err).(reader.ErrLowTimeout); !ok {
		t.Errorf("want (reader.ErrLowTimeout), got (%v)", err)
	}

	err = reader.SetTimeout(time.Millisecond * 10)(&r)
	if _, ok := errors.Cause(err).(reader.ErrLowTimeout); !ok {
		t.Errorf("want (reader.ErrLowTimeout), got (%v)", err)
	}

	err = reader.SetTimeout(time.Second)(&r)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}

func TestSetBackoff(t *testing.T) {
	r := reader_testing.Reader{}
	err := reader.SetBackoff(4)(&r)
	if _, ok := errors.Cause(err).(reader.ErrLowBackoffValue); !ok {
		t.Errorf("want (reader.ErrLowBackoffValue), got (%v)", err)
	}

	err = reader.SetBackoff(5)(&r)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}
