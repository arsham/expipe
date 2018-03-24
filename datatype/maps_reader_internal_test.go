// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/antonholmquist/jason"
	"github.com/arsham/expipe/internal"
	"github.com/spf13/viper"
)

func TestLoadMapsReaderGCTypes(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")

	input := bytes.NewBuffer([]byte(`
    gc_types:
        PauseEnd
        memstats.PauseNs
    `))

	v.ReadConfig(input)
	maps := MapsFromViper(v)
	for _, c := range []string{"PauseEnd", "memstats.PauseNs"} {
		if !internal.StringInSlice(c, maps.gcTypes) {
			t.Errorf("(%s) not found in returned valued. got (%s)", c, strings.Join(maps.gcTypes, ", "))
		}
	}
	input = bytes.NewBuffer([]byte(`
    gc_types:
    `))

	v.ReadConfig(input)
	maps = MapsFromViper(v)
	if len(maps.gcTypes) != 0 {
		t.Fatalf("expected empty results, got (%v)", maps.gcTypes)
	}

	input = bytes.NewBuffer([]byte(``))

	v.ReadConfig(input)
	maps = MapsFromViper(v)
	if len(maps.gcTypes) != 0 {
		t.Fatalf("expected empty results, got (%v)", maps.gcTypes)
	}
}

func TestLoadMapsReaderMemoryTypes(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")

	tc := map[string]string{
		"alloc":     "mb",
		"sys":       "gb",
		"heapalloc": "mb",
		"heapsys":   "mb",
	}
	input := bytes.NewBuffer([]byte(`
    memory_bytes:
        Alloc: mb
        Sys: gb
        HeapAlloc: mb
        HeapSys: mb
    `))
	v.ReadConfig(input)

	maps := MapsFromViper(v)
	var returnedNames []string

	for name := range maps.memoryTypes {
		returnedNames = append(returnedNames, string(name))
	}

	for _, name := range []string{"alloc", "sys", "heapalloc", "heapsys"} {
		if !stringInMapKeys(name, maps.memoryTypes) {
			t.Errorf("(%s) not found in returned valued. got (%s)", name, strings.Join(returnedNames, ", "))
		}
		if tc[name] != string(maps.memoryTypes[name]) {
			t.Errorf("want (%s), got (%s)", tc[name], string(maps.memoryTypes[name]))
		}
	}

	input = bytes.NewBuffer([]byte(`
    memory_bytes:
    `))
	v.ReadConfig(input)
	maps = MapsFromViper(v)
	if len(maps.memoryTypes) != 0 {
		t.Fatalf("expected empty results, got (%v)", maps.memoryTypes)
	}

	input = bytes.NewBuffer([]byte(``))
	v.ReadConfig(input)
	maps = MapsFromViper(v)
	if len(maps.memoryTypes) != 0 {
		t.Fatalf("expected empty results, got (%v)", maps.memoryTypes)
	}
}

func TestGetArrayValue(t *testing.T) {
	t.Parallel()
	prefix := "Mr. "
	name := "Devil"
	m := &MapConvert{}

	expected := &FloatListType{Key: "Mr. Devil", Value: []float64{}}
	result := m.getArrayValue(prefix, name, []*jason.Value{})
	if !result.Equal(expected) {
		t.Errorf("want (%v), got (%v)", expected, result)
	}

	str, _ := jason.NewValueFromBytes([]byte(`{"sdss":"sdfs"}`))
	result = m.getArrayValue(prefix, name, []*jason.Value{str})
	if result != nil {
		t.Errorf("want nil, got (%v)", result)
	}
}

func TestMapCopy(t *testing.T) {
	t.Parallel()
	m := &MapConvert{
		gcTypes:     []string{"first"},
		memoryTypes: map[string]memType{"second": "third"},
	}
	c := m.Copy()
	cc, ok := c.(*MapConvert)
	if !ok {
		t.Fatalf("want MapConvert, got (%T)", c)
	}
	if m == c {
		t.Error("wasn't copied")
	}
	if !reflect.DeepEqual(cc, m) {
		t.Fatalf("want (%v), got (%v)", m, c)
	}
}

func TestMapConvertMockCopy(t *testing.T) {
	t.Parallel()
	m := &MapConvertMock{
		GCTypes:     []string{"first"},
		MemoryTypes: map[string]MemTypeMock{"second": {"third"}},
	}
	c := m.Copy()
	cc, ok := c.(*MapConvertMock)
	if !ok {
		t.Fatalf("want MapConvertMock, got (%T)", c)
	}
	if m == c {
		t.Error("wasn't copied")
	}
	if !reflect.DeepEqual(cc, m) {
		t.Fatalf("want (%v), got (%v)", m, c)
	}
}
