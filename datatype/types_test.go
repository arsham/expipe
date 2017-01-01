// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/arsham/expvastic/datatype"
)

func TestGetQueryString(t *testing.T) {
	now := time.Now()
	tStr := fmt.Sprintf(`"@timestamp":"%s"`, now.Format("2006-01-02T15:04:05.999999-07:00"))

	testCase := []struct {
		input    []datatype.DataType
		expected string
	}{
		{
			[]datatype.DataType{},
			fmt.Sprintf("{%s}", tStr),
		},
		{
			[]datatype.DataType{&datatype.FloatType{Key: "test", Value: 3.4}},
			fmt.Sprintf(`{%s,"test":%f}`, tStr, 3.4),
		},
		{
			[]datatype.DataType{&datatype.StringType{Key: "test", Value: "3a"}, &datatype.FloatType{Key: "test2", Value: 2.2}},
			fmt.Sprintf(`{%s,"test":"%s","test2":%f}`, tStr, "3a", 2.2),
		},
		{
			[]datatype.DataType{&datatype.StringType{Key: "test2", Value: "3a"}, &datatype.KiloByteType{Key: "test3", Value: 3.3}},
			fmt.Sprintf(`{%s,"test2":"%s","test3":%f}`, tStr, "3a", 3.3/1024.0),
		},
	}

	for i, tc := range testCase {
		name := fmt.Sprintf("case %d", i)
		t.Run(name, func(t *testing.T) {
			contaner := datatype.NewContainer(tc.input)
			results := contaner.Bytes(now)
			if !reflect.DeepEqual(results, []byte(tc.expected)) {
				t.Errorf("want (%s) got (%s)", tc.expected, results)
			}
		})
	}
}

func TestDataTypeEquality(t *testing.T) {
	// FloatListType
	// GCListType
	type inputType struct {
		a datatype.DataType
		b datatype.DataType
	}
	testCase := []struct {
		input    inputType
		expected bool
	}{
		{input: inputType{a: &datatype.FloatType{Key: "a", Value: 1.1}, b: &datatype.FloatType{Key: "a", Value: 1.1}}, expected: true},  // 0
		{input: inputType{a: &datatype.FloatType{Key: "a", Value: 1.1}, b: &datatype.FloatType{Key: "b", Value: 1.1}}, expected: false}, // 1
		{input: inputType{a: &datatype.FloatType{Key: "a", Value: 1.1}, b: &datatype.FloatType{Key: "a", Value: 1.2}}, expected: false}, // 2

		{input: inputType{a: &datatype.StringType{Key: "a", Value: "1.1"}, b: &datatype.StringType{Key: "a", Value: "1.2"}}, expected: false}, // 3
		{input: inputType{a: &datatype.StringType{Key: "a", Value: "1.1"}, b: &datatype.StringType{Key: "b", Value: "1.1"}}, expected: false}, // 4
		{input: inputType{a: &datatype.StringType{Key: "a", Value: "1.1"}, b: &datatype.StringType{Key: "a", Value: "1.2"}}, expected: false}, // 5

		{input: inputType{a: &datatype.ByteType{Key: "a", Value: 1.1}, b: &datatype.ByteType{Key: "a", Value: 1.2}}, expected: false}, // 6
		{input: inputType{a: &datatype.ByteType{Key: "a", Value: 1.1}, b: &datatype.ByteType{Key: "b", Value: 1.1}}, expected: false}, // 7
		{input: inputType{a: &datatype.ByteType{Key: "a", Value: 1.1}, b: &datatype.ByteType{Key: "a", Value: 1.2}}, expected: false}, // 8

		{input: inputType{a: &datatype.KiloByteType{Key: "a", Value: 1.1}, b: &datatype.KiloByteType{Key: "a", Value: 1.2}}, expected: false}, // 9
		{input: inputType{a: &datatype.KiloByteType{Key: "a", Value: 1.1}, b: &datatype.KiloByteType{Key: "b", Value: 1.1}}, expected: false}, // 10
		{input: inputType{a: &datatype.KiloByteType{Key: "a", Value: 1.1}, b: &datatype.KiloByteType{Key: "a", Value: 1.2}}, expected: false}, // 11

		{input: inputType{a: &datatype.MegaByteType{Key: "a", Value: 1.1}, b: &datatype.MegaByteType{Key: "a", Value: 1.2}}, expected: false}, // 12
		{input: inputType{a: &datatype.MegaByteType{Key: "a", Value: 1.1}, b: &datatype.MegaByteType{Key: "b", Value: 1.1}}, expected: false}, // 13
		{input: inputType{a: &datatype.MegaByteType{Key: "a", Value: 1.1}, b: &datatype.MegaByteType{Key: "a", Value: 1.2}}, expected: false}, // 14

		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}, b: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}}, expected: true},            // 15
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}, b: &datatype.FloatListType{Key: "b", Value: []float64{1.1}}}, expected: false},           // 16
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}, b: &datatype.FloatListType{Key: "a", Value: []float64{1.2}}}, expected: false},           // 17
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1, 2.2}}, b: &datatype.FloatListType{Key: "a", Value: []float64{1.1, 2.2}}}, expected: true},  // 18
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1, 2.2}}, b: &datatype.FloatListType{Key: "a", Value: []float64{2.2, 1.1}}}, expected: true},  // 19
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}, b: &datatype.FloatListType{Key: "b", Value: []float64{1.1}}}, expected: false},           // 20
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}, b: &datatype.FloatListType{Key: "a", Value: []float64{1.2}}}, expected: false},           // 21
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1, 2.2}}, b: &datatype.FloatListType{Key: "b", Value: []float64{2.2, 1.1}}}, expected: false}, // 22
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1, 2.2}}, b: &datatype.FloatListType{Key: "b", Value: []float64{1.1, 2.2}}}, expected: false}, // 23

		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1}}, b: &datatype.GCListType{Key: "a", Value: []uint64{1}}}, expected: true},        // 24
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1}}, b: &datatype.GCListType{Key: "b", Value: []uint64{1}}}, expected: false},       // 25
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1, 2}}, b: &datatype.GCListType{Key: "a", Value: []uint64{1, 2}}}, expected: true},  // 26
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1, 2}}, b: &datatype.GCListType{Key: "a", Value: []uint64{2, 1}}}, expected: true},  // 27
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1}}, b: &datatype.GCListType{Key: "b", Value: []uint64{1}}}, expected: false},       // 28
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1, 2}}, b: &datatype.GCListType{Key: "b", Value: []uint64{2, 1}}}, expected: false}, // 29
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1, 2}}, b: &datatype.GCListType{Key: "b", Value: []uint64{1, 2}}}, expected: false}, // 30
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1}}, b: &datatype.GCListType{Key: "b", Value: []uint64{1, 2}}}, expected: false},    // 30
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1, 2}}, b: &datatype.GCListType{Key: "b", Value: []uint64{1}}}, expected: false},    // 30

		{input: inputType{a: &datatype.FloatType{Key: "a", Value: 1.1}, b: &datatype.StringType{Key: "a", Value: "1.1"}}, expected: false},                      // 31
		{input: inputType{a: &datatype.StringType{Key: "a", Value: "1.1"}, b: &datatype.FloatType{Key: "a", Value: 1.2}}, expected: false},                      // 32
		{input: inputType{a: &datatype.ByteType{Key: "a", Value: 1.1}, b: &datatype.KiloByteType{Key: "a", Value: 1.2}}, expected: false},                       // 33
		{input: inputType{a: &datatype.KiloByteType{Key: "a", Value: 1.1}, b: &datatype.MegaByteType{Key: "a", Value: 1.2}}, expected: false},                   // 34
		{input: inputType{a: &datatype.MegaByteType{Key: "a", Value: 1.1}, b: &datatype.ByteType{Key: "a", Value: 1.2}}, expected: false},                       // 35
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}, b: &datatype.GCListType{Key: "a", Value: []uint64{1}}}, expected: false}, // 36
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1}}, b: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}}, expected: false}, // 37
	}

	for i, tc := range testCase {
		name := fmt.Sprintf("case %d", i)
		t.Run(name, func(t *testing.T) {
			res := tc.input.a.Equal(tc.input.b)
			if res != tc.expected {
				t.Errorf("want (%t) got (%t)", tc.expected, res)
			}
		})
	}
}
