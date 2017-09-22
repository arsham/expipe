// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/arsham/expvastic/internal/datatype"
)

func TestGetByteRepresentation(t *testing.T) {
	now := time.Now()
	tStr := fmt.Sprintf(`"@timestamp":"%s"`, now.Format("2006-01-02T15:04:05.999999-07:00"))

	testCase := []struct {
		input    []datatype.DataType
		expected string
	}{
		{ // 0
			[]datatype.DataType{},
			fmt.Sprintf("{%s}", tStr),
		},
		{ // 1
			[]datatype.DataType{&datatype.FloatType{Key: "test", Value: 3.4}},
			fmt.Sprintf(`{%s,"test":%f}`, tStr, 3.4),
		},
		{ // 2
			[]datatype.DataType{&datatype.StringType{Key: "test", Value: "3.4"}},
			fmt.Sprintf(`{%s,"test":"%s"}`, tStr, "3.4"),
		},
		{ // 3
			[]datatype.DataType{&datatype.ByteType{Key: "test", Value: 1024 * 1024 * 2}},
			fmt.Sprintf(`{%s,"test":%f}`, tStr, 2.0),
		},
		{ // 4
			[]datatype.DataType{&datatype.MegaByteType{Key: "test", Value: 1024 * 1024 * 3}},
			fmt.Sprintf(`{%s,"test":%f}`, tStr, 3.0),
		},
		{ // 5
			[]datatype.DataType{&datatype.StringType{Key: "test", Value: "3a"}, &datatype.FloatType{Key: "test2", Value: 2.2}},
			fmt.Sprintf(`{%s,"test":"%s","test2":%f}`, tStr, "3a", 2.2),
		},
		{ // 6
			[]datatype.DataType{&datatype.StringType{Key: "test2", Value: "3a"}, &datatype.KiloByteType{Key: "test3", Value: 3.3}},
			fmt.Sprintf(`{%s,"test2":"%s","test3":%f}`, tStr, "3a", 3.3/1024.0),
		},
		{ // 7
			[]datatype.DataType{&datatype.StringType{Key: "test", Value: "3a"}, &datatype.FloatListType{Key: "test2", Value: []float64{1.1, 2.2}}},
			fmt.Sprintf(`{%s,"test":"%s","test2":[%f,%f]}`, tStr, "3a", 1.1, 2.2),
		},
		{ // 8
			[]datatype.DataType{&datatype.FloatType{Key: "test", Value: 1.1}, &datatype.GCListType{Key: "test2", Value: []uint64{100, 10}}},
			fmt.Sprintf(`{%s,"test":%f,"test2":[%d,%d]}`, tStr, 1.1, 0, 0),
		},
		{ // 9
			[]datatype.DataType{&datatype.FloatType{Key: "test", Value: 1.1}, &datatype.GCListType{Key: "test2", Value: []uint64{1000, 2000}}},
			fmt.Sprintf(`{%s,"test":%f,"test2":[%d,%d]}`, tStr, 1.1, 1, 2),
		},
	}

	for i, tc := range testCase {
		name := fmt.Sprintf("case %d", i)
		t.Run(name, func(t *testing.T) {
			contaner := datatype.New(tc.input)
			results := contaner.Bytes(now)
			if !reflect.DeepEqual(results, []byte(tc.expected)) {
				t.Errorf("want (%s) got (%s)", tc.expected, results)
			}
		})
	}
}

func TestDataTypeEquality(t *testing.T) {
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
		{input: inputType{a: &datatype.FloatType{Key: "a", Value: 1.1}, b: nil}, expected: false},                                       // 3

		{input: inputType{a: &datatype.StringType{Key: "a", Value: "1.1"}, b: &datatype.StringType{Key: "a", Value: "1.1"}}, expected: true},  // 4
		{input: inputType{a: &datatype.StringType{Key: "a", Value: "1.1"}, b: &datatype.StringType{Key: "a", Value: "1.2"}}, expected: false}, // 5
		{input: inputType{a: &datatype.StringType{Key: "a", Value: "1.1"}, b: &datatype.StringType{Key: "b", Value: "1.1"}}, expected: false}, // 6
		{input: inputType{a: &datatype.StringType{Key: "a", Value: "1.1"}, b: &datatype.StringType{Key: "a", Value: "1.2"}}, expected: false}, // 7
		{input: inputType{a: &datatype.StringType{Key: "a", Value: "1.1"}, b: nil}, expected: false},                                          // 8

		{input: inputType{a: &datatype.ByteType{Key: "a", Value: 1.1}, b: &datatype.ByteType{Key: "a", Value: 1.1}}, expected: true},  // 9
		{input: inputType{a: &datatype.ByteType{Key: "a", Value: 1.1}, b: &datatype.ByteType{Key: "a", Value: 1.2}}, expected: false}, // 10
		{input: inputType{a: &datatype.ByteType{Key: "a", Value: 1.1}, b: &datatype.ByteType{Key: "b", Value: 1.1}}, expected: false}, // 11
		{input: inputType{a: &datatype.ByteType{Key: "a", Value: 1.1}, b: &datatype.ByteType{Key: "a", Value: 1.2}}, expected: false}, // 12
		{input: inputType{a: &datatype.ByteType{Key: "a", Value: 1.1}, b: nil}, expected: false},                                      // 13

		{input: inputType{a: &datatype.KiloByteType{Key: "a", Value: 1.1}, b: &datatype.KiloByteType{Key: "a", Value: 1.1}}, expected: true},  // 14
		{input: inputType{a: &datatype.KiloByteType{Key: "a", Value: 1.1}, b: &datatype.KiloByteType{Key: "a", Value: 1.2}}, expected: false}, // 15
		{input: inputType{a: &datatype.KiloByteType{Key: "a", Value: 1.1}, b: &datatype.KiloByteType{Key: "b", Value: 1.1}}, expected: false}, // 16
		{input: inputType{a: &datatype.KiloByteType{Key: "a", Value: 1.1}, b: &datatype.KiloByteType{Key: "a", Value: 1.2}}, expected: false}, // 17
		{input: inputType{a: &datatype.KiloByteType{Key: "a", Value: 1.1}, b: nil}, expected: false},                                          // 18

		{input: inputType{a: &datatype.MegaByteType{Key: "a", Value: 1.1}, b: &datatype.MegaByteType{Key: "a", Value: 1.1}}, expected: true},  // 19
		{input: inputType{a: &datatype.MegaByteType{Key: "a", Value: 1.1}, b: &datatype.MegaByteType{Key: "a", Value: 1.2}}, expected: false}, // 20
		{input: inputType{a: &datatype.MegaByteType{Key: "a", Value: 1.1}, b: &datatype.MegaByteType{Key: "a", Value: 1.2}}, expected: false}, // 21
		{input: inputType{a: &datatype.MegaByteType{Key: "a", Value: 1.1}, b: &datatype.MegaByteType{Key: "b", Value: 1.1}}, expected: false}, // 22
		{input: inputType{a: &datatype.MegaByteType{Key: "a", Value: 1.1}, b: nil}, expected: false},                                          // 23

		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}, b: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}}, expected: true},            // 24
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}, b: &datatype.FloatListType{Key: "b", Value: []float64{1.1}}}, expected: false},           // 25
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}, b: &datatype.FloatListType{Key: "a", Value: []float64{1.2}}}, expected: false},           // 26
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1, 2.2}}, b: &datatype.FloatListType{Key: "a", Value: []float64{1.1, 2.2}}}, expected: true},  // 27
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1, 2.2}}, b: &datatype.FloatListType{Key: "a", Value: []float64{2.2, 1.1}}}, expected: true},  // 28
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}, b: &datatype.FloatListType{Key: "b", Value: []float64{1.1}}}, expected: false},           // 29
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}, b: &datatype.FloatListType{Key: "a", Value: []float64{1.2}}}, expected: false},           // 30
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1, 2.2}}, b: &datatype.FloatListType{Key: "b", Value: []float64{2.2, 1.1}}}, expected: false}, // 31
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1, 2.2}}, b: &datatype.FloatListType{Key: "b", Value: []float64{1.1, 2.2}}}, expected: false}, // 32
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1, 2.2}}, b: nil}, expected: false},                                                           // 33

		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1}}, b: &datatype.GCListType{Key: "a", Value: []uint64{1}}}, expected: true},        // 34
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1}}, b: &datatype.GCListType{Key: "b", Value: []uint64{1}}}, expected: false},       // 35
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1, 2}}, b: &datatype.GCListType{Key: "a", Value: []uint64{1, 2}}}, expected: true},  // 36
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1, 2}}, b: &datatype.GCListType{Key: "a", Value: []uint64{2, 1}}}, expected: true},  // 37
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1}}, b: &datatype.GCListType{Key: "b", Value: []uint64{1}}}, expected: false},       // 38
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1, 2}}, b: &datatype.GCListType{Key: "b", Value: []uint64{2, 1}}}, expected: false}, // 39
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1, 2}}, b: &datatype.GCListType{Key: "b", Value: []uint64{1, 2}}}, expected: false}, // 40
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1}}, b: &datatype.GCListType{Key: "b", Value: []uint64{1, 2}}}, expected: false},    // 41
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1, 2}}, b: &datatype.GCListType{Key: "b", Value: []uint64{1}}}, expected: false},    // 42
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1, 2}}, b: nil}, expected: false},                                                   // 43

		{input: inputType{a: &datatype.FloatType{Key: "a", Value: 1.1}, b: &datatype.StringType{Key: "a", Value: "1.1"}}, expected: false},                      // 44
		{input: inputType{a: &datatype.StringType{Key: "a", Value: "1.1"}, b: &datatype.FloatType{Key: "a", Value: 1.2}}, expected: false},                      // 45
		{input: inputType{a: &datatype.ByteType{Key: "a", Value: 1.1}, b: &datatype.KiloByteType{Key: "a", Value: 1.2}}, expected: false},                       // 46
		{input: inputType{a: &datatype.KiloByteType{Key: "a", Value: 1.1}, b: &datatype.MegaByteType{Key: "a", Value: 1.2}}, expected: false},                   // 47
		{input: inputType{a: &datatype.MegaByteType{Key: "a", Value: 1.1}, b: &datatype.ByteType{Key: "a", Value: 1.2}}, expected: false},                       // 48
		{input: inputType{a: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}, b: &datatype.GCListType{Key: "a", Value: []uint64{1}}}, expected: false}, // 49
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1}}, b: &datatype.FloatListType{Key: "a", Value: []float64{1.1}}}, expected: false}, // 50
		{input: inputType{a: &datatype.GCListType{Key: "a", Value: []uint64{1}}, b: nil}, expected: false},                                                      // 51
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
