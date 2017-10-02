// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype_test

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"github.com/antonholmquist/jason"
	"github.com/arsham/expipe/internal/datatype"
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
		{"byte_value", "12", datatype.ByteType{"byte_value", 12}},
		{"alloc", strconv.Itoa(1 * datatype.KILOBYTE), datatype.KiloByteType{"alloc", 1 * datatype.KILOBYTE}},
		{"sys", strconv.Itoa(12 * datatype.MEGABYTE), datatype.MegaByteType{"sys", 12 * datatype.MEGABYTE}},
		{"not_provided", `"anything"`, datatype.StringType{"not_provided", "anything"}},
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
					t.Errorf("want (%#v), got (%#v)", tc.expected, value)
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
		{"float_list", `[]`, &datatype.FloatListType{"float_list", []float64{}}},
		{"float_list", `[0.1,1.2,2.3,3.4,666]`, &datatype.FloatListType{"float_list", []float64{0.1, 1.2, 2.3, 3.4, 666}}},
		{"float_list", `[0.1,1.2,2.3,3.4,666]`, &datatype.FloatListType{"float_list", []float64{2.3, 3.4, 666, 0.1, 1.2}}},
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
					t.Errorf("want (%#v), got (%#v)", tc.expected, value)
				}
			}
		})
	}
}

// Make sure the memstats.PauseNs is not overwritten by PauseNs
func TestNestedPauseNsRegression(t *testing.T) {
	t.Parallel()
	// input := bytes.NewBuffer([]byte(`{"memstats": {"Alloc":6865888,"TotalAlloc":14509024, "PauseNs":[438238,506913]}}`))
	input := bytes.NewBuffer([]byte(`{"memstats": {"PauseNs":[438238,506913]}}`))
	expected := &datatype.GCListType{Key: "memstats.PauseNs", Value: []uint64{438238, 506913}}
	mapper := datatype.DefaultMapper()
	container := datatype.JobResultDataTypes(input.Bytes(), mapper)
	if !container.List()[0].Equal(expected) {
		t.Errorf("want (%#v), got (%#v)", expected, container.List()[0])
	}
}
