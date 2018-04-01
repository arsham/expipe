// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
	"reflect"
	"testing"

	"github.com/antonholmquist/jason"
)

func TestGetArrayValue(t *testing.T) {
	t.Parallel()
	prefix := "Mr. "
	name := "Devil"
	m := &MapConvert{}

	expected := &FloatListType{Key: "Mr. Devil", Value: []float64{}}
	result := m.arrayValue(prefix, name, []*jason.Value{})
	if !result.Equal(expected) {
		t.Errorf("result.Equal(expected) = false, result = (%v); want (%v)", result, expected)
	}

	str, _ := jason.NewValueFromBytes([]byte(`{"sdss":"sdfs"}`))
	result = m.arrayValue(prefix, name, []*jason.Value{str})
	if result != nil {
		t.Errorf("result = (%v); want (nil)", result)
	}
}

func TestGetMemoryTypes(t *testing.T) {
	jb, err := jason.NewValueFromBytes([]byte(`6.5`))
	if err != nil {
		t.Fatalf("NewValueFromBytes(): err = (%#v); want (nil)", err)
	}
	jkb, err := jason.NewValueFromBytes([]byte(`6.5`))
	if err != nil {
		t.Fatalf("NewValueFromBytes(): err = (%#v); want (nil)", err)
	}
	jmb, err := jason.NewValueFromBytes([]byte(`6.5`))
	if err != nil {
		t.Fatalf("NewValueFromBytes(): err = (%#v); want (nil)", err)
	}
	memType := map[string]string{
		"b":  "b",
		"kb": "kb",
		"mb": "mb",
	}
	tcs := []struct {
		tcName string
		name   string
		j      *jason.Value
		dt     DataType
		ok     bool
	}{
		{"ByteType", "B", jb, &ByteType{}, true},
		{"KiloByteType", "KB", jkb, &KiloByteType{}, true},
		{"MegaByteType", "MB", jmb, &MegaByteType{}, true},
		{"NoneType", "something else", jmb, nil, false},
	}
	m := &MapConvert{
		MemoryTypes: memType,
	}
	for _, tc := range tcs {
		t.Run(tc.tcName, func(t *testing.T) {
			dt, ok := m.getMemoryTypes("", tc.name, tc.j)
			if reflect.TypeOf(dt) != reflect.TypeOf(tc.dt) {
				t.Errorf("MapConvert.getMemoryTypes() dt is (%v), want (%v)", dt, tc.dt)
			}
			if ok != tc.ok {
				t.Errorf("MapConvert.getMemoryTypes() ok = (%v), want (%v)", ok, tc.ok)
			}
		})
	}
}
