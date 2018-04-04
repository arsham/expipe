// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype_test

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"testing/quick"

	"github.com/antonholmquist/jason"
	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/tools"
	"github.com/spf13/viper"
)

func TestGetMemoryTypeValues(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")

	tcs := []struct {
		name     string
		value    string
		expected datatype.DataType
	}{
		{
			"byte_value",
			"12",
			datatype.NewByteType("byte_value", 12),
		},
		{
			"alloc",
			strconv.Itoa(1 * datatype.KiloByte),
			datatype.NewKiloByteType("alloc", 1*datatype.KiloByte),
		},
		{
			"sys",
			strconv.Itoa(12 * datatype.MegaByte),
			datatype.NewMegaByteType("sys", 12*datatype.MegaByte),
		},
		{
			"not_provided",
			`"anything"`,
			datatype.NewStringType("not_provided", "anything"),
		},
	}

	input := bytes.NewBuffer([]byte(`
    memory_bytes:
        byte_value: b
        Alloc: kb
        Sys: mb
    `))
	v.ReadConfig(input)
	maps := datatype.MapsFromViper(v)
	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			v, _ := jason.NewValueFromBytes([]byte(tc.value))
			results := maps.Values("", map[string]*jason.Value{tc.name: v})
			for _, value := range results {
				if !tc.expected.Equal(value) {
					t.Errorf("tc.expected.Equal(value): value = (%#v); want (%#v)", value, tc.expected)
				}
			}
		})
	}
}

func TestGetFloatListValues(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")

	tcs := []struct {
		name     string
		value    string
		expected datatype.DataType
	}{
		{
			"float_list",
			`[]`,
			datatype.NewFloatListType("float_list", []float64{}),
		},
		{
			"float_list",
			`[0.1,1.2,2.3,3.4,666]`,
			datatype.NewFloatListType("float_list", []float64{0.1, 1.2, 2.3, 3.4, 666}),
		},
		{
			"float_list",
			`[0.1,1.2,2.3,3.4,666]`,
			datatype.NewFloatListType("float_list", []float64{2.3, 3.4, 666, 0.1, 1.2}),
		},
	}

	input := bytes.NewBuffer([]byte(``))
	v.ReadConfig(input)
	maps := datatype.MapsFromViper(v)
	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			v, _ := jason.NewValueFromBytes([]byte(tc.value))
			results := maps.Values("", map[string]*jason.Value{tc.name: v})
			for _, value := range results {
				if !tc.expected.Equal(value) {
					t.Errorf("tc.expected.Equal(value): value = (%#v); want (%#v)", value, tc.expected)
				}
			}
		})
	}
}

// Make sure the memstats.PauseNs is not overwritten by PauseNs
func TestNestedPauseNsRegression(t *testing.T) {
	t.Parallel()
	input := bytes.NewBuffer([]byte(`{"memstats": {"PauseNs":[438238,506913]}}`))
	expected := &datatype.GCListType{Key: "memstats.PauseNs", Value: []uint64{438238, 506913}}
	mapper := datatype.DefaultMapper()
	container, _ := datatype.JobResultDataTypes(input.Bytes(), mapper)
	if !container.List()[0].Equal(expected) {
		t.Errorf("container.List()[0] = (%#v); want (%#v)", container.List()[0], expected)
	}
}

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
	maps := datatype.MapsFromViper(v)
	for _, c := range []string{"PauseEnd", "memstats.PauseNs"} {
		if !tools.StringInSlice(c, maps.GCTypes) {
			v := strings.Join(maps.GCTypes, ", ")
			t.Errorf("tools.StringInSlice(c, maps.GCTypes): (%s) not found in returned valued. got (%s)", c, v)
		}
	}
	input = bytes.NewBuffer([]byte(`
    gc_types:
    `))

	v.ReadConfig(input)
	maps = datatype.MapsFromViper(v)
	if len(maps.GCTypes) != 0 {
		t.Fatalf("len(maps.GCTypes) = (%v); want empty results", maps.GCTypes)
	}

	input = bytes.NewBuffer([]byte(``))

	v.ReadConfig(input)
	maps = datatype.MapsFromViper(v)
	if len(maps.GCTypes) != 0 {
		t.Fatalf("len(maps.GCTypes) = (%v); want empty results", maps.GCTypes)
	}
}

func TestLoadMapsReaderMemoryTypes(t *testing.T) {
	t.Parallel()
	var returnedNames []string
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
	maps := datatype.MapsFromViper(v)
	for name := range maps.MemoryTypes {
		returnedNames = append(returnedNames, string(name))
	}

	for _, name := range []string{"alloc", "sys", "heapalloc", "heapsys"} {
		if _, found := maps.MemoryTypes[strings.ToLower(name)]; !found {
			t.Errorf("(%s) not found in returned valued. got (%s)", name, strings.Join(returnedNames, ", "))
		}
		if tc[name] != string(maps.MemoryTypes[name]) {
			t.Errorf("tc[name] = (%s); want (%s)", string(maps.MemoryTypes[name]), tc[name])
		}
	}

	input = bytes.NewBuffer([]byte(`
    memory_bytes:
    `))
	v.ReadConfig(input)
	maps = datatype.MapsFromViper(v)
	if len(maps.MemoryTypes) != 0 {
		t.Fatalf("len(maps.MemoryTypes) = (%v); want (empty results)", maps.MemoryTypes)
	}

	input = bytes.NewBuffer([]byte(``))
	v.ReadConfig(input)
	maps = datatype.MapsFromViper(v)
	if len(maps.MemoryTypes) != 0 {
		t.Fatalf("len(maps.MemoryTypes) = (%v); want (empty results)", maps.MemoryTypes)
	}
}

func TestMapConvertMockCopy(t *testing.T) {
	t.Parallel()
	f := func(gcTypes []string, memTypes map[string]string) bool {
		m := &datatype.MapConvertMock{
			GCTypes:     gcTypes,
			MemoryTypes: memTypes,
		}
		c := m.Copy()
		cc, ok := c.(*datatype.MapConvertMock)
		if !ok {
			t.Errorf("c.(*datatype.MapConvertMock): c = (%T); want (*datatype.MapConvertMock)", c)
			return false
		}
		if m == c {
			t.Error("m.Copy(): wasn't copied")
			return false
		}
		if !reflect.DeepEqual(cc, m) {
			t.Errorf("reflect.DeepEqual(cc, m): c = (%v); want (%v)", c, m)
			return false
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestMapCopy(t *testing.T) {
	t.Parallel()
	f := func(gcTypes []string, memTypes map[string]string) bool {
		m := &datatype.MapConvert{
			GCTypes:     gcTypes,
			MemoryTypes: memTypes,
		}
		c := m.Copy()
		cc, ok := c.(*datatype.MapConvert)
		if !ok {
			t.Errorf("c.(*datatype.MapConvert) = (%T); want (datatype.MapConvert)", c)
			return false
		}
		if m == c {
			t.Error("m.Copy(): wasn't copied")
			return false
		}
		if !reflect.DeepEqual(cc, m) {
			t.Errorf("reflect.DeepEqual(cc, m): c = (%v); want (%v)", c, m)
			return false
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
