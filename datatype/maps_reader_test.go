// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
    "bytes"
    "fmt"
    "strconv"
    "strings"
    "testing"

    "github.com/antonholmquist/jason"
    "github.com/arsham/expvastic/lib"
    "github.com/spf13/viper"
)

func TestLoadMapsReaderGCTypes(t *testing.T) {
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
        if !lib.StringInSlice(c, maps.gcTypes) {
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
        t.Fatalf("expected empty results, got (%d)", maps.memoryTypes)
    }

    input = bytes.NewBuffer([]byte(``))
    v.ReadConfig(input)
    maps = MapsFromViper(v)
    if len(maps.memoryTypes) != 0 {
        t.Fatalf("expected empty results, got (%d)", maps.memoryTypes)
    }
}

func TestGetMemoryTypeValues(t *testing.T) {

    v := viper.New()
    v.SetConfigType("yaml")

    tcs := []struct {
        name     string
        value    string
        expected DataType
    }{
        {"byte_value", "12", ByteType{"byte_value", 12}},
        {"alloc", strconv.Itoa(1 * KILOBYTE), KiloByteType{"alloc", 1 * KILOBYTE}},
        {"sys", strconv.Itoa(12 * MEGABYTE), MegaByteType{"sys", 12 * MEGABYTE}},
        {"not_provided", `"anything"`, StringType{"not_provided", "anything"}},
    }

    input := bytes.NewBuffer([]byte(`
    memory_bytes:
        byte_value: b
        Alloc: kb
        Sys: mb
    `))
    v.ReadConfig(input)
    maps := MapsFromViper(v)
    for i, tc := range tcs {
        name := fmt.Sprintf("case_%d", i)
        t.Run(name, func(t *testing.T) {
            v, _ := jason.NewValueFromBytes([]byte(tc.value))
            vchan := maps.Values("", map[string]*jason.Value{tc.name: v})
            if value := <-vchan; !tc.expected.Equal(value) {
                t.Errorf("want (%#v), got (%#v)", tc.expected, value)
            }
        })
    }
}

func TestGetFloatListValues(t *testing.T) {
    v := viper.New()
    v.SetConfigType("yaml")

    tcs := []struct {
        name     string
        value    string
        expected DataType
    }{
        {"float_list", `[]`, &FloatListType{"float_list", []float64{}}},
        {"float_list", `[0.1,1.2,2.3,3.4,666]`, &FloatListType{"float_list", []float64{0.1, 1.2, 2.3, 3.4, 666}}},
        {"float_list", `[0.1,1.2,2.3,3.4,666]`, &FloatListType{"float_list", []float64{2.3, 3.4, 666, 0.1, 1.2}}},
    }

    input := bytes.NewBuffer([]byte(``))
    v.ReadConfig(input)
    maps := MapsFromViper(v)
    for i, tc := range tcs {
        name := fmt.Sprintf("case_%d", i)
        t.Run(name, func(t *testing.T) {
            v, _ := jason.NewValueFromBytes([]byte(tc.value))
            vchan := maps.Values("", map[string]*jason.Value{tc.name: v})
            if value := <-vchan; !tc.expected.Equal(value) {
                t.Errorf("want (%#v), got (%#v)", tc.expected, value)
            }
        })
    }
}

// Make sure the memstats.PauseNs is not overwritten by PauseNs
func TestNestedPauseNsRegression(t *testing.T) {
    // input := bytes.NewBuffer([]byte(`{"memstats": {"Alloc":6865888,"TotalAlloc":14509024, "PauseNs":[438238,506913]}}`))
    input := bytes.NewBuffer([]byte(`{"memstats": {"PauseNs":[438238,506913]}}`))
    expected := &GCListType{Key: "memstats.PauseNs", Value: []uint64{438238, 506913}}
    mapper := DefaultMapper()
    container := JobResultDataTypes(input, mapper)
    if !container.List()[0].Equal(expected) {
        t.Errorf("want (%#v), got (%#v)", expected, container.List()[0])
    }
}