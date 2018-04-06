// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype_test

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/antonholmquist/jason"
	"github.com/arsham/expipe/datatype"
)

type caseType struct {
	name   string
	prefix string
	value  []byte
	want   []datatype.DataType
}

func TestGetByteRepresentation(t *testing.T) {
	now := time.Now()
	tStr := fmt.Sprintf(`"@timestamp":"%s"`, now.Format("2006-01-02T15:04:05.999999-07:00"))

	testCase := []struct {
		name     string
		input    []datatype.DataType
		expected string
	}{
		{
			name:     "0",
			input:    []datatype.DataType{},
			expected: fmt.Sprintf("{%s}", tStr),
		},
		{
			name:     "1",
			input:    []datatype.DataType{datatype.NewFloatType("test", 3.4)},
			expected: fmt.Sprintf(`{%s,"test":%f}`, tStr, 3.4),
		},
		{
			name:     "2",
			input:    []datatype.DataType{datatype.NewStringType("test", "3.4")},
			expected: fmt.Sprintf(`{%s,"test":"%s"}`, tStr, "3.4"),
		},
		{
			name:     "3",
			input:    []datatype.DataType{datatype.NewByteType("test", 1024*1024*2)},
			expected: fmt.Sprintf(`{%s,"test":%f}`, tStr, 2.0),
		},
		{
			name:     "4",
			input:    []datatype.DataType{datatype.NewMegaByteType("test", 1024*1024*3)},
			expected: fmt.Sprintf(`{%s,"test":%f}`, tStr, 3.0),
		},
		{
			name: "5",
			input: []datatype.DataType{
				datatype.NewStringType("test", "3a"),
				datatype.NewFloatType("test2", 2.2),
			},
			expected: fmt.Sprintf(`{%s,"test":"%s","test2":%f}`, tStr, "3a", 2.2),
		},
		{
			name: "6",
			input: []datatype.DataType{
				datatype.NewStringType("test2", "3a"),
				datatype.NewKiloByteType("test3", 3.3),
			},
			expected: fmt.Sprintf(`{%s,"test2":"%s","test3":%f}`, tStr, "3a", 3.3/1024.0),
		},
		{
			name: "7",
			input: []datatype.DataType{
				datatype.NewStringType("test", "3a"),
				datatype.NewFloatListType("test2", []float64{1.1, 2.2}),
			},
			expected: fmt.Sprintf(`{%s,"test":"%s","test2":[%f,%f]}`, tStr, "3a", 1.1, 2.2),
		},
		{
			name: "8",
			input: []datatype.DataType{
				datatype.NewFloatType("test", 1.1),
				datatype.NewGCListType("test2", []uint64{100, 10}),
			},
			expected: fmt.Sprintf(`{%s,"test":%f,"test2":[%d,%d]}`, tStr, 1.1, 0, 0),
		},
		{
			name: "9",
			input: []datatype.DataType{
				datatype.NewFloatType("test", 1.1),
				datatype.NewGCListType("test2", []uint64{1000, 2000}),
			},
			expected: fmt.Sprintf(`{%s,"test":%f,"test2":[%d,%d]}`, tStr, 1.1, 1, 2),
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			contaner := datatype.New(tc.input)
			results := new(bytes.Buffer)
			n, err := contaner.Generate(results, now)
			if err != nil {
				t.Errorf("Generate(results, now): err = (%v); want (nil)", err)
			}
			if n != len(tc.expected) {
				t.Errorf("n = (%d); want (%d)", n, len(tc.expected))
			}
			if !reflect.DeepEqual(results.String(), tc.expected) {
				t.Errorf("DeepEqual(results, tc.expected): results = (%s); want (%s)",
					results.String(),
					tc.expected,
				)
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
		number   int
		input    inputType
		expected bool
	}{
		{number: 0, input: inputType{a: datatype.NewFloatType("a", 1.1), b: datatype.NewFloatType("a", 1.1)}, expected: true},
		{number: 1, input: inputType{a: datatype.NewFloatType("a", 1.1), b: datatype.NewFloatType("b", 1.1)}, expected: false},
		{number: 2, input: inputType{a: datatype.NewFloatType("a", 1.1), b: datatype.NewFloatType("a", 1.2)}, expected: false},
		{number: 3, input: inputType{a: datatype.NewFloatType("a", 1.1), b: nil}, expected: false},

		{number: 4, input: inputType{a: datatype.NewStringType("a", "1.1"), b: datatype.NewStringType("a", "1.1")}, expected: true},
		{number: 5, input: inputType{a: datatype.NewStringType("a", "1.1"), b: datatype.NewStringType("a", "1.2")}, expected: false},
		{number: 6, input: inputType{a: datatype.NewStringType("a", "1.1"), b: datatype.NewStringType("b", "1.1")}, expected: false},
		{number: 7, input: inputType{a: datatype.NewStringType("a", "1.1"), b: datatype.NewStringType("a", "1.2")}, expected: false},
		{number: 8, input: inputType{a: datatype.NewStringType("a", "1.1"), b: nil}, expected: false},

		{number: 9, input: inputType{a: datatype.NewByteType("a", 1.1), b: datatype.NewByteType("a", 1.1)}, expected: true},
		{number: 10, input: inputType{a: datatype.NewByteType("a", 1.1), b: datatype.NewByteType("a", 1.2)}, expected: false},
		{number: 11, input: inputType{a: datatype.NewByteType("a", 1.1), b: datatype.NewByteType("b", 1.1)}, expected: false},
		{number: 12, input: inputType{a: datatype.NewByteType("a", 1.1), b: datatype.NewByteType("a", 1.2)}, expected: false},
		{number: 13, input: inputType{a: datatype.NewByteType("a", 1.1), b: nil}, expected: false},

		{number: 14, input: inputType{a: datatype.NewKiloByteType("a", 1.1), b: datatype.NewKiloByteType("a", 1.1)}, expected: true},
		{number: 15, input: inputType{a: datatype.NewKiloByteType("a", 1.1), b: datatype.NewKiloByteType("a", 1.2)}, expected: false},
		{number: 16, input: inputType{a: datatype.NewKiloByteType("a", 1.1), b: datatype.NewKiloByteType("b", 1.1)}, expected: false},
		{number: 17, input: inputType{a: datatype.NewKiloByteType("a", 1.1), b: datatype.NewKiloByteType("a", 1.2)}, expected: false},
		{number: 18, input: inputType{a: datatype.NewKiloByteType("a", 1.1), b: nil}, expected: false},

		{number: 19, input: inputType{a: datatype.NewMegaByteType("a", 1.1), b: datatype.NewMegaByteType("a", 1.1)}, expected: true},
		{number: 20, input: inputType{a: datatype.NewMegaByteType("a", 1.1), b: datatype.NewMegaByteType("a", 1.2)}, expected: false},
		{number: 21, input: inputType{a: datatype.NewMegaByteType("a", 1.1), b: datatype.NewMegaByteType("a", 1.2)}, expected: false},
		{number: 22, input: inputType{a: datatype.NewMegaByteType("a", 1.1), b: datatype.NewMegaByteType("b", 1.1)}, expected: false},
		{number: 23, input: inputType{a: datatype.NewMegaByteType("a", 1.1), b: nil}, expected: false},

		{number: 24, input: inputType{a: datatype.NewFloatListType("a", []float64{1.1}), b: datatype.NewFloatListType("a", []float64{1.1})}, expected: true},
		{number: 25, input: inputType{a: datatype.NewFloatListType("a", []float64{1.1}), b: datatype.NewFloatListType("b", []float64{1.1})}, expected: false},
		{number: 26, input: inputType{a: datatype.NewFloatListType("a", []float64{1.1}), b: datatype.NewFloatListType("a", []float64{1.2})}, expected: false},
		{number: 27, input: inputType{a: datatype.NewFloatListType("a", []float64{1.1, 2.2}), b: datatype.NewFloatListType("a", []float64{1.1, 2.2})}, expected: true},
		{number: 28, input: inputType{a: datatype.NewFloatListType("a", []float64{1.1, 2.2}), b: datatype.NewFloatListType("a", []float64{2.2, 1.1})}, expected: true},
		{number: 29, input: inputType{a: datatype.NewFloatListType("a", []float64{1.1}), b: datatype.NewFloatListType("b", []float64{1.1})}, expected: false},
		{number: 30, input: inputType{a: datatype.NewFloatListType("a", []float64{1.1}), b: datatype.NewFloatListType("a", []float64{1.2})}, expected: false},
		{number: 31, input: inputType{a: datatype.NewFloatListType("a", []float64{1.1, 2.2}), b: datatype.NewFloatListType("b", []float64{2.2, 1.1})}, expected: false},
		{number: 32, input: inputType{a: datatype.NewFloatListType("a", []float64{1.1, 2.2}), b: datatype.NewFloatListType("b", []float64{1.1, 2.2})}, expected: false},
		{number: 33, input: inputType{a: datatype.NewFloatListType("a", []float64{1.1, 2.2}), b: nil}, expected: false},

		{number: 34, input: inputType{a: datatype.NewGCListType("a", []uint64{1}), b: datatype.NewGCListType("a", []uint64{1})}, expected: true},
		{number: 35, input: inputType{a: datatype.NewGCListType("a", []uint64{1}), b: datatype.NewGCListType("b", []uint64{1})}, expected: false},
		{number: 36, input: inputType{a: datatype.NewGCListType("a", []uint64{1, 2}), b: datatype.NewGCListType("a", []uint64{1, 2})}, expected: true},
		{number: 37, input: inputType{a: datatype.NewGCListType("a", []uint64{1, 2}), b: datatype.NewGCListType("a", []uint64{2, 1})}, expected: true},
		{number: 38, input: inputType{a: datatype.NewGCListType("a", []uint64{1}), b: datatype.NewGCListType("b", []uint64{1})}, expected: false},
		{number: 39, input: inputType{a: datatype.NewGCListType("a", []uint64{1, 2}), b: datatype.NewGCListType("b", []uint64{2, 1})}, expected: false},
		{number: 40, input: inputType{a: datatype.NewGCListType("a", []uint64{1, 2}), b: datatype.NewGCListType("b", []uint64{1, 2})}, expected: false},
		{number: 41, input: inputType{a: datatype.NewGCListType("a", []uint64{1}), b: datatype.NewGCListType("b", []uint64{1, 2})}, expected: false},
		{number: 42, input: inputType{a: datatype.NewGCListType("a", []uint64{1, 2}), b: datatype.NewGCListType("b", []uint64{1})}, expected: false},
		{number: 43, input: inputType{a: datatype.NewGCListType("a", []uint64{1, 2}), b: nil}, expected: false},
		{number: 43, input: inputType{a: datatype.NewGCListType("a", []uint64{1, 2}), b: datatype.NewGCListType("a", []uint64{2, 1, 3})}, expected: false},

		{number: 44, input: inputType{a: datatype.NewFloatType("a", 1.1), b: datatype.NewStringType("a", "1.1")}, expected: false},
		{number: 45, input: inputType{a: datatype.NewStringType("a", "1.1"), b: datatype.NewFloatType("a", 1.2)}, expected: false},
		{number: 46, input: inputType{a: datatype.NewByteType("a", 1.1), b: datatype.NewKiloByteType("a", 1.2)}, expected: false},
		{number: 47, input: inputType{a: datatype.NewKiloByteType("a", 1.1), b: datatype.NewMegaByteType("a", 1.2)}, expected: false},
		{number: 48, input: inputType{a: datatype.NewMegaByteType("a", 1.1), b: datatype.NewByteType("a", 1.2)}, expected: false},
		{number: 49, input: inputType{a: datatype.NewFloatListType("a", []float64{1.1}), b: datatype.NewGCListType("a", []uint64{1})}, expected: false},
		{number: 50, input: inputType{a: datatype.NewGCListType("a", []uint64{1}), b: datatype.NewFloatListType("a", []float64{1.1})}, expected: false},
		{number: 51, input: inputType{a: datatype.NewGCListType("a", []uint64{1}), b: nil}, expected: false},
	}

	for _, tc := range testCase {
		name := fmt.Sprintf("case %d", tc.number)
		t.Run(name, func(t *testing.T) {
			res := tc.input.a.Equal(tc.input.b)
			if res != tc.expected {
				t.Errorf("res = (%t); want (%t)", res, tc.expected)
			}
		})
	}
}

func TestGetJasonValues(t *testing.T) {
	t.Parallel()
	mapper := &datatype.MapConvertMock{}
	for i, tc := range testCase() {
		name := fmt.Sprintf("case %d", i)
		t.Run(name, func(t *testing.T) {
			var payload []datatype.DataType
			obj, _ := jason.NewObjectFromBytes(tc.value)
			payload = append(payload, mapper.Values(tc.prefix, obj.Map())...)

			if len(payload) == 0 {
				t.Errorf("len(payload) = (%d); want (%d)", len(payload), len(tc.want))
				return
			}
			results := datatype.New(payload)
			if !isIn(results.List(), tc.want) {
				t.Errorf("isIn(List(), tc.want): List() = (%v); want (%v)", results.List(), tc.want)
			}
		})
	}
}

func TestGetJasonValuesAddToContainer(t *testing.T) {
	t.Parallel()
	mapper := &datatype.MapConvertMock{}
	for i, tc := range testCase() {
		name := fmt.Sprintf("case %d", i)
		var container datatype.Container
		t.Run(name, func(t *testing.T) {
			obj, _ := jason.NewObjectFromBytes(tc.value)
			for _, value := range mapper.Values(tc.prefix, obj.Map()) {
				container.Add(value)
			}
			if !isIn(container.List(), tc.want) {
				t.Errorf("isIn(List(), tc.expected): List() = (%#v); want (%#v)",
					container.List(),
					tc.want,
				)
			}
		})
	}
}

func TestFromReader(t *testing.T) {
	t.Parallel()
	mapper := &datatype.MapConvertMock{}
	for i, tc := range testCase() {
		if tc.prefix != "" {
			continue
		}
		name := fmt.Sprintf("case %d", i)
		t.Run(name, func(t *testing.T) {
			results, err := datatype.JobResultDataTypes(tc.value, mapper)
			if err != nil {
				t.Errorf("err = (%s); want (nil)", err)
			}
			if !isIn(results.List(), tc.want) {
				t.Errorf("isIn(List(), tc.expected): List() = (%s); want (%s)", results.List(), tc.want)
			}
		})
	}
}

func isIn(a, b []datatype.DataType) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !inArray(a[i], b) {
			return false
		}
	}
	return true
}
func TestIsIn(t *testing.T) {
	a := datatype.NewStringType("key", "value")
	aa := datatype.NewStringType("key", "value")
	b := datatype.NewFloatType("key", 6.66)
	c := datatype.NewFloatType("key2", 6.66)

	tcs := []struct {
		name   string
		left   []datatype.DataType
		right  []datatype.DataType
		result bool
	}{
		{"a in nothing", []datatype.DataType{a}, []datatype.DataType{}, false},
		{"a in aa", []datatype.DataType{a}, []datatype.DataType{aa}, true},
		{"a in b", []datatype.DataType{a}, []datatype.DataType{b}, false},
		{"ab in ba", []datatype.DataType{a, b}, []datatype.DataType{b, a}, true},
		{"abc in bca", []datatype.DataType{a, b, c}, []datatype.DataType{b, c, a}, true},
		{"ab in bca", []datatype.DataType{a, b}, []datatype.DataType{b, c, a}, false},
		{"bca in ab", []datatype.DataType{b, c, a}, []datatype.DataType{a, b}, false},
		{"abc in ab", []datatype.DataType{a, b, c}, []datatype.DataType{a, b}, false},
		{"aab in baa", []datatype.DataType{a, a, b}, []datatype.DataType{b, a, a}, true},
		{"aab in aba", []datatype.DataType{a, a, b}, []datatype.DataType{b, a, a}, true},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			r := isIn(tc.left, tc.right)
			if r != tc.result {
				t.Errorf("isIn(tc.left, tc.right) = (%t); want (%t)", r, tc.result)
			}
		})
	}
}

func testCase() []caseType {
	return []caseType{
		{
			name:   "0",
			prefix: "",
			value:  []byte(`{"FloatType": 123.4}`),
			want:   []datatype.DataType{datatype.NewFloatType("FloatType", 123.4)},
		},
		{
			name:   "1",
			prefix: "",
			value:  []byte(`{"StringType": "Random: 666"}`),
			want:   []datatype.DataType{datatype.NewStringType("StringType", "Random: 666")},
		},
		{
			name:   "2",
			prefix: "aaa.",
			value:  []byte(`{"Prefixed": 666.777}`),
			want:   []datatype.DataType{datatype.NewFloatType("aaa.Prefixed", 666.777)},
		},
		{
			name:   "3",
			prefix: "",
			value:  []byte(`{"Nested": {"FloatType": 666.777}}`),
			want:   []datatype.DataType{datatype.NewFloatType("Nested.FloatType", 666.777)},
		},
		{
			name:   "4",
			prefix: "",
			value:  []byte(`{"Multy": 666.77, "Nested": {"FloatType": 666.999}}`),
			want:   []datatype.DataType{datatype.NewFloatType("Multy", 666.77), datatype.NewFloatType("Nested.FloatType", 666.999)},
		},
		{
			name:   "5",
			prefix: "",
			value:  []byte(`{"Multy": 666.77, "Nested": {"FloatType": 666.999}}`),
			want:   []datatype.DataType{datatype.NewFloatType("Nested.FloatType", 666.999), datatype.NewFloatType("Multy", 666.77)},
		},
		{
			name:   "6",
			prefix: "",
			value:  []byte(`{"FloatListType": []}`),
			want:   []datatype.DataType{datatype.NewFloatListType("FloatListType", []float64{})},
		},
		{
			name:   "7",
			prefix: "",
			value:  []byte(`{"FloatListType": [0.1,1.2,2.3,3.4,666]}`),
			want:   []datatype.DataType{datatype.NewFloatListType("FloatListType", []float64{0.1, 1.2, 2.3, 3.4, 666})},
		},
		{
			name:   "8",
			prefix: "",
			value:  []byte(`{"PauseNs": []}`),
			want:   []datatype.DataType{datatype.NewGCListType("PauseNs", []uint64{})},
		},
		{
			name:   "9",
			prefix: "",
			value:  []byte(`{"PauseNs": [0,0,0,0,12481868021080215863,1481868005672005459,1481868012773129951,666000,11481937182104993300]}`),
			want: []datatype.DataType{datatype.NewGCListType(
				"PauseNs",
				[]uint64{12481868021080215863, 1481868005672005459, 1481868012773129951, 666000, 11481937182104993300},
			)},
		},
		{
			name:   "10",
			prefix: "",
			value:  []byte(`{"TotalAlloc": 0}`),
			want:   []datatype.DataType{datatype.NewByteType("TotalAlloc", 0)},
		},
		{
			name:   "11",
			prefix: "",
			value:  []byte(`{"TotalAlloc": 236478234}`),
			want:   []datatype.DataType{datatype.NewByteType("TotalAlloc", 236478234)},
		},
		{
			name:   "12",
			prefix: "",
			value:  []byte(`{"PauseNs": [1481938891973801922,1481938893974355168,1481938895974915920,1481938897975467569,1481938899975919573,1481938901976464855,1481938903977051088,1481938905977636658,1481938907978221684,1481938909978619244,1481938911979100042,1481938913979740815,1481938915980232455,1481938917980671611,1481938919981183393,1481938921981827241,1481938923982308276,1481938925982865139,1481938927983327577,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}`),
			want: []datatype.DataType{datatype.NewGCListType(
				"PauseNs",
				[]uint64{1481938891973801922, 1481938893974355168, 1481938895974915920, 1481938897975467569, 1481938899975919573, 1481938901976464855, 1481938903977051088, 1481938905977636658, 1481938907978221684, 1481938909978619244, 1481938911979100042, 1481938913979740815, 1481938915980232455, 1481938917980671611, 1481938919981183393, 1481938921981827241, 1481938923982308276, 1481938925982865139, 1481938927983327577},
			)},
		},
		{
			name:   "13",
			prefix: "",
			value:  []byte(`{"memstats": {"Alloc":2780496,"TotalAlloc": 236478234}}`),
			want: []datatype.DataType{
				datatype.NewByteType("memstats.Alloc", 2780496),
				datatype.NewByteType("memstats.TotalAlloc", 236478234),
			},
		},
	}
}

func TestReset(t *testing.T) {
	t.Parallel()
	tcs := []datatype.DataType{
		datatype.NewFloatType("HahyTcVWS", 666.6),
		datatype.NewStringType("SLVRNGMPdMm", "IzWQtIqPGESks"),
		datatype.NewFloatListType("YliEXjPfyL", []float64{2452.4, 245245.44, 245554.23, 454.555}),
		datatype.NewGCListType("nVuXyTztBucw", []uint64{234, 535, 133255, 36563, 242544, 3563534}),
		datatype.NewByteType("NkHeYtrYqjnJJJnMB", 666.2),
		datatype.NewKiloByteType("AVCbQeWMAfdvWRugZJ", 4463436.3),
		datatype.NewMegaByteType("zrZvbdxQCzIJZZ", 3343453.345),
	}
	for _, tc := range tcs {
		p := make([]byte, 10)
		buf := new(bytes.Buffer)
		n1, err := buf.ReadFrom(tc)
		if err != nil {
			t.Errorf("err = (%#v); want (nil)", err)
			continue
		}
		n, err := tc.Read(p)
		if err != io.EOF {
			t.Errorf("Read(): err = (%#v); want (%#v)", err, io.EOF)
		}
		if n != 0 {
			t.Errorf("Read(): n = (%#v); want (0)", n)
		}
		tc.Reset()
		n2, err := buf.ReadFrom(tc)
		if err != nil {
			t.Errorf("err = (%#v); want (nil)", err)
		}
		if n2 != n1 {
			t.Errorf("n2 = (%d); want (%d)", n2, n1)
		}
	}
}
