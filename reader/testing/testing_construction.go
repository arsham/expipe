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

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/tools"
	"github.com/pkg/errors"
)

const (
	name     = "the name"
	typeName = "my type"
)

func shouldNotChangeTheInput(t testing.TB, cons Constructor) {
	endpoint := cons.TestServer().URL
	interval := time.Second
	timeout := time.Second
	logger := tools.DiscardLogger()
	cons.SetName(name)
	cons.SetTypeName(typeName)
	cons.SetEndpoint(endpoint)
	cons.SetInterval(interval)
	cons.SetTimeout(timeout)
	cons.SetLogger(logger)
	red, err := cons.Object()
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if red.Name() != name {
		t.Errorf("red.Name() = (%s); want (%s)", red.Name(), name)
	}
	if red.TypeName() != typeName {
		t.Errorf("red.TypeName() = (%s); want (%s)", red.TypeName(), typeName)
	}
	if red.Interval() != interval {
		t.Errorf("red.Interval() = (%s); want (%s)", red.Interval().String(), interval.String())
	}
	if red.Timeout() != timeout {
		t.Errorf("red.Timeout() = (%d); want (%d)", red.Timeout(), timeout)
	}
	if red.Endpoint() != endpoint {
		t.Errorf("red.Endpoint() = (%s); want (%s)", red.Endpoint(), endpoint)
	}
}

func nameCheck(t testing.TB, cons Constructor) {
	cons.SetTypeName(typeName)
	cons.SetTimeout(time.Hour)
	cons.SetEndpoint(cons.TestServer().URL)
	red, err := cons.Object()
	if errors.Cause(err) != reader.ErrEmptyName {
		t.Errorf("err = (%v); want (reader.ErrEmptyName)", err)
	}
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("red = (%v); want (nil)", red)
	}

	cons.SetName("")
	red, err = cons.Object()
	if errors.Cause(err) != reader.ErrEmptyName {
		t.Errorf("err = (%v); want (reader.ErrEmptyName)", err)
	}
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("red = (%v); want (nil)", red)
	}
}

func typeNameCheck(t testing.TB, cons Constructor) {
	cons.SetName(name)
	cons.SetTimeout(time.Hour)
	cons.SetEndpoint(cons.TestServer().URL)
	red, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	if reflect.ValueOf(red).IsNil() {
		t.Fatal("red = (nil); want (DataReader)")
	}
	if red.TypeName() != red.Name() {
		t.Errorf("TypeName() = (%s); want (%s)", red.TypeName(), red.Name())
	}

	cons.SetTypeName("")
	red, err = cons.Object()
	if errors.Cause(err) != reader.ErrEmptyTypeName {
		t.Errorf("err = (%#v); want (reader.ErrEmptyTypeName)", err)
	}
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("red = (%v); want (nil)", red)
	}
}

func intervalCheck(t testing.TB, cons Constructor) {
	endpoint := cons.TestServer().URL
	interval := 0
	cons.SetEndpoint(endpoint)
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetInterval(time.Duration(interval))
	red, err := cons.Object()
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("red = (%v); want (nil)", red)
	}
	if _, ok := errors.Cause(err).(reader.LowIntervalError); !ok {
		t.Errorf("err.(reader.LowIntervalError) = (%#v); want (reader.LowIntervalError)", err)
	}
	if !strings.Contains(err.Error(), strconv.Itoa(interval)) {
		t.Errorf("Contains(err.Error(), interval): want (%s) to be in (%s)", strconv.Itoa(interval), err.Error())
	}
}

func endpointCheck(t testing.TB, cons Constructor) {
	var ok bool
	invalidEndpoint := "this is invalid"
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetTimeout(time.Second)
	cons.SetInterval(time.Second)

	red, err := cons.Object()
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("red = (%#v); want (nil)", red)
	}
	if errors.Cause(err) != reader.ErrEmptyEndpoint {
		t.Errorf("err = (%#v); want (reader.ErrEmptyEndpoint)", err)
	}

	cons.SetEndpoint("")
	red, err = cons.Object()
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("red = (%v); want (nil)", red)
	}
	if errors.Cause(err) != reader.ErrEmptyEndpoint {
		t.Errorf("err = (%#v); want (reader.ErrEmptyEndpoint)", err)
	}

	cons.SetEndpoint(invalidEndpoint)
	red, err = cons.Object()
	if !reflect.ValueOf(red).IsNil() {
		t.Errorf("red = (%v); want (nil)", red)
	}
	if _, ok = errors.Cause(err).(reader.InvalidEndpointError); !ok {
		t.Errorf("err.(reader.InvalidEndpointError) = (%#v); want (reader.InvalidEndpointError)", err)
	}
	if ok && !strings.Contains(err.Error(), invalidEndpoint) {
		t.Errorf("want (%s) to be in (%s)", invalidEndpoint, err.Error())
	}
}

func timeoutCheck(t testing.TB, cons Constructor) {
	cons.SetName(name)
	cons.SetEndpoint(cons.TestServer().URL)
	red, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	if reflect.ValueOf(red).IsNil() {
		t.Fatal("red = (nil); want (DataReader)")
	}
	if red.Timeout() != 5*time.Second {
		t.Errorf("Timeout() = (%s); want (%s)", red.Timeout(), 5*time.Second)
	}
}

func setMapperCheck(t testing.TB, cons Constructor) {
	mapper := datatype.DefaultMapper()
	cons.SetName(name)
	cons.SetEndpoint(cons.TestServer().URL)
	cons.SetMapper(mapper)
	cons.SetLogger(tools.DiscardLogger())
	red, err := cons.Object()
	if errors.Cause(err) != nil {
		t.Errorf("err = (%#v); want (nil)", err)
	}
	if reflect.ValueOf(red).IsNil() {
		t.Fatal("red = (nil); want (DataReader)")
	}
	if red.Mapper() != mapper {
		t.Errorf("Mapper() = (%#v); want (%#v)", red.Mapper(), mapper)
	}
}
