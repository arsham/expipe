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
	Prefix string
	Value  []byte
	Want   []datatype.DataType
}

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
			[]datatype.DataType{datatype.NewFloatType("test", 3.4)},
			fmt.Sprintf(`{%s,"test":%f}`, tStr, 3.4),
		},
		{ // 2
			[]datatype.DataType{datatype.NewStringType("test", "3.4")},
			fmt.Sprintf(`{%s,"test":"%s"}`, tStr, "3.4"),
		},
		{ // 3
			[]datatype.DataType{datatype.NewByteType("test", 1024*1024*2)},
			fmt.Sprintf(`{%s,"test":%f}`, tStr, 2.0),
		},
		{ // 4
			[]datatype.DataType{datatype.NewMegaByteType("test", 1024*1024*3)},
			fmt.Sprintf(`{%s,"test":%f}`, tStr, 3.0),
		},
		{ // 5
			[]datatype.DataType{datatype.NewStringType("test", "3a"), datatype.NewFloatType("test2", 2.2)},
			fmt.Sprintf(`{%s,"test":"%s","test2":%f}`, tStr, "3a", 2.2),
		},
		{ // 6
			[]datatype.DataType{datatype.NewStringType("test2", "3a"), datatype.NewKiloByteType("test3", 3.3)},
			fmt.Sprintf(`{%s,"test2":"%s","test3":%f}`, tStr, "3a", 3.3/1024.0),
		},
		{ // 7
			[]datatype.DataType{datatype.NewStringType("test", "3a"), datatype.NewFloatListType("test2", []float64{1.1, 2.2})},
			fmt.Sprintf(`{%s,"test":"%s","test2":[%f,%f]}`, tStr, "3a", 1.1, 2.2),
		},
		{ // 8
			[]datatype.DataType{datatype.NewFloatType("test", 1.1), datatype.NewGCListType("test2", []uint64{100, 10})},
			fmt.Sprintf(`{%s,"test":%f,"test2":[%d,%d]}`, tStr, 1.1, 0, 0),
		},
		{ // 9
			[]datatype.DataType{datatype.NewFloatType("test", 1.1), datatype.NewGCListType("test2", []uint64{1000, 2000})},
			fmt.Sprintf(`{%s,"test":%f,"test2":[%d,%d]}`, tStr, 1.1, 1, 2),
		},
	}

	for i, tc := range testCase {
		name := fmt.Sprintf("case %d", i)
		t.Run(name, func(t *testing.T) {
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
				t.Errorf("DeepEqual(results.String(), tc.expected): results = (%s); want (%s)", results.String(), tc.expected)
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
		{input: inputType{a: datatype.NewFloatType("a", 1.1), b: datatype.NewFloatType("a", 1.1)}, expected: true},  // 0
		{input: inputType{a: datatype.NewFloatType("a", 1.1), b: datatype.NewFloatType("b", 1.1)}, expected: false}, // 1
		{input: inputType{a: datatype.NewFloatType("a", 1.1), b: datatype.NewFloatType("a", 1.2)}, expected: false}, // 2
		{input: inputType{a: datatype.NewFloatType("a", 1.1), b: nil}, expected: false},                             // 3

		{input: inputType{a: datatype.NewStringType("a", "1.1"), b: datatype.NewStringType("a", "1.1")}, expected: true},  // 4
		{input: inputType{a: datatype.NewStringType("a", "1.1"), b: datatype.NewStringType("a", "1.2")}, expected: false}, // 5
		{input: inputType{a: datatype.NewStringType("a", "1.1"), b: datatype.NewStringType("b", "1.1")}, expected: false}, // 6
		{input: inputType{a: datatype.NewStringType("a", "1.1"), b: datatype.NewStringType("a", "1.2")}, expected: false}, // 7
		{input: inputType{a: datatype.NewStringType("a", "1.1"), b: nil}, expected: false},                                // 8

		{input: inputType{a: datatype.NewByteType("a", 1.1), b: datatype.NewByteType("a", 1.1)}, expected: true},  // 9
		{input: inputType{a: datatype.NewByteType("a", 1.1), b: datatype.NewByteType("a", 1.2)}, expected: false}, // 10
		{input: inputType{a: datatype.NewByteType("a", 1.1), b: datatype.NewByteType("b", 1.1)}, expected: false}, // 11
		{input: inputType{a: datatype.NewByteType("a", 1.1), b: datatype.NewByteType("a", 1.2)}, expected: false}, // 12
		{input: inputType{a: datatype.NewByteType("a", 1.1), b: nil}, expected: false},                            // 13

		{input: inputType{a: datatype.NewKiloByteType("a", 1.1), b: datatype.NewKiloByteType("a", 1.1)}, expected: true},  // 14
		{input: inputType{a: datatype.NewKiloByteType("a", 1.1), b: datatype.NewKiloByteType("a", 1.2)}, expected: false}, // 15
		{input: inputType{a: datatype.NewKiloByteType("a", 1.1), b: datatype.NewKiloByteType("b", 1.1)}, expected: false}, // 16
		{input: inputType{a: datatype.NewKiloByteType("a", 1.1), b: datatype.NewKiloByteType("a", 1.2)}, expected: false}, // 17
		{input: inputType{a: datatype.NewKiloByteType("a", 1.1), b: nil}, expected: false},                                // 18

		{input: inputType{a: datatype.NewMegaByteType("a", 1.1), b: datatype.NewMegaByteType("a", 1.1)}, expected: true},  // 19
		{input: inputType{a: datatype.NewMegaByteType("a", 1.1), b: datatype.NewMegaByteType("a", 1.2)}, expected: false}, // 20
		{input: inputType{a: datatype.NewMegaByteType("a", 1.1), b: datatype.NewMegaByteType("a", 1.2)}, expected: false}, // 21
		{input: inputType{a: datatype.NewMegaByteType("a", 1.1), b: datatype.NewMegaByteType("b", 1.1)}, expected: false}, // 22
		{input: inputType{a: datatype.NewMegaByteType("a", 1.1), b: nil}, expected: false},                                // 23

		{input: inputType{a: datatype.NewFloatListType("a", []float64{1.1}), b: datatype.NewFloatListType("a", []float64{1.1})}, expected: true},            // 24
		{input: inputType{a: datatype.NewFloatListType("a", []float64{1.1}), b: datatype.NewFloatListType("b", []float64{1.1})}, expected: false},           // 25
		{input: inputType{a: datatype.NewFloatListType("a", []float64{1.1}), b: datatype.NewFloatListType("a", []float64{1.2})}, expected: false},           // 26
		{input: inputType{a: datatype.NewFloatListType("a", []float64{1.1, 2.2}), b: datatype.NewFloatListType("a", []float64{1.1, 2.2})}, expected: true},  // 27
		{input: inputType{a: datatype.NewFloatListType("a", []float64{1.1, 2.2}), b: datatype.NewFloatListType("a", []float64{2.2, 1.1})}, expected: true},  // 28
		{input: inputType{a: datatype.NewFloatListType("a", []float64{1.1}), b: datatype.NewFloatListType("b", []float64{1.1})}, expected: false},           // 29
		{input: inputType{a: datatype.NewFloatListType("a", []float64{1.1}), b: datatype.NewFloatListType("a", []float64{1.2})}, expected: false},           // 30
		{input: inputType{a: datatype.NewFloatListType("a", []float64{1.1, 2.2}), b: datatype.NewFloatListType("b", []float64{2.2, 1.1})}, expected: false}, // 31
		{input: inputType{a: datatype.NewFloatListType("a", []float64{1.1, 2.2}), b: datatype.NewFloatListType("b", []float64{1.1, 2.2})}, expected: false}, // 32
		{input: inputType{a: datatype.NewFloatListType("a", []float64{1.1, 2.2}), b: nil}, expected: false},                                                 // 33

		{input: inputType{a: datatype.NewGCListType("a", []uint64{1}), b: datatype.NewGCListType("a", []uint64{1})}, expected: true},        // 34
		{input: inputType{a: datatype.NewGCListType("a", []uint64{1}), b: datatype.NewGCListType("b", []uint64{1})}, expected: false},       // 35
		{input: inputType{a: datatype.NewGCListType("a", []uint64{1, 2}), b: datatype.NewGCListType("a", []uint64{1, 2})}, expected: true},  // 36
		{input: inputType{a: datatype.NewGCListType("a", []uint64{1, 2}), b: datatype.NewGCListType("a", []uint64{2, 1})}, expected: true},  // 37
		{input: inputType{a: datatype.NewGCListType("a", []uint64{1}), b: datatype.NewGCListType("b", []uint64{1})}, expected: false},       // 38
		{input: inputType{a: datatype.NewGCListType("a", []uint64{1, 2}), b: datatype.NewGCListType("b", []uint64{2, 1})}, expected: false}, // 39
		{input: inputType{a: datatype.NewGCListType("a", []uint64{1, 2}), b: datatype.NewGCListType("b", []uint64{1, 2})}, expected: false}, // 40
		{input: inputType{a: datatype.NewGCListType("a", []uint64{1}), b: datatype.NewGCListType("b", []uint64{1, 2})}, expected: false},    // 41
		{input: inputType{a: datatype.NewGCListType("a", []uint64{1, 2}), b: datatype.NewGCListType("b", []uint64{1})}, expected: false},    // 42
		{input: inputType{a: datatype.NewGCListType("a", []uint64{1, 2}), b: nil}, expected: false},                                         // 43

		{input: inputType{a: datatype.NewFloatType("a", 1.1), b: datatype.NewStringType("a", "1.1")}, expected: false},                      // 44
		{input: inputType{a: datatype.NewStringType("a", "1.1"), b: datatype.NewFloatType("a", 1.2)}, expected: false},                      // 45
		{input: inputType{a: datatype.NewByteType("a", 1.1), b: datatype.NewKiloByteType("a", 1.2)}, expected: false},                       // 46
		{input: inputType{a: datatype.NewKiloByteType("a", 1.1), b: datatype.NewMegaByteType("a", 1.2)}, expected: false},                   // 47
		{input: inputType{a: datatype.NewMegaByteType("a", 1.1), b: datatype.NewByteType("a", 1.2)}, expected: false},                       // 48
		{input: inputType{a: datatype.NewFloatListType("a", []float64{1.1}), b: datatype.NewGCListType("a", []uint64{1})}, expected: false}, // 49
		{input: inputType{a: datatype.NewGCListType("a", []uint64{1}), b: datatype.NewFloatListType("a", []float64{1.1})}, expected: false}, // 50
		{input: inputType{a: datatype.NewGCListType("a", []uint64{1}), b: nil}, expected: false},                                            // 51
	}

	for i, tc := range testCase {
		name := fmt.Sprintf("case %d", i)
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
			obj, _ := jason.NewObjectFromBytes(tc.Value)
			payload = append(payload, mapper.Values(tc.Prefix, obj.Map())...)

			if len(payload) == 0 {
				t.Errorf("len(payload) = (%d); want (%d)", len(payload), len(tc.Want))
				return
			}
			results := datatype.New(payload)
			if !isIn(results.List(), tc.Want) {
				t.Errorf("isIn(results.List(), tc.Want): results.List() = (%v); want (%v)", results.List(), tc.Want)
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
			obj, _ := jason.NewObjectFromBytes(tc.Value)
			for _, value := range mapper.Values(tc.Prefix, obj.Map()) {
				container.Add(value)
			}
			if !isIn(container.List(), tc.Want) {
				t.Errorf("isIn(container.List(), tc.expected): container.List() = (%#v); want (%#v)", container.List(), tc.Want)
			}
		})
	}
}

func TestFromReader(t *testing.T) {
	t.Parallel()
	mapper := &datatype.MapConvertMock{}
	for i, tc := range testCase() {
		if tc.Prefix != "" {
			continue
		}
		name := fmt.Sprintf("case %d", i)
		t.Run(name, func(t *testing.T) {
			results, err := datatype.JobResultDataTypes(tc.Value, mapper)
			if err != nil {
				t.Errorf("err = (%s); want (nil)", err)
			}
			if !isIn(results.List(), tc.Want) {
				t.Errorf("isIn(results.List(), tc.expected): results.List() = (%s); want (%s)", results.List(), tc.Want)
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
		{ //0
			"",
			[]byte(`{"FloatType": 123.4}`),
			[]datatype.DataType{datatype.NewFloatType("FloatType", 123.4)},
		},
		{ //1
			"",
			[]byte(`{"StringType": "Random: 666"}`),
			[]datatype.DataType{datatype.NewStringType("StringType", "Random: 666")},
		},
		{ //2
			"aaa.",
			[]byte(`{"Prefixed": 666.777}`),
			[]datatype.DataType{datatype.NewFloatType("aaa.Prefixed", 666.777)},
		},
		{ //3
			"",
			[]byte(`{"Nested": {"FloatType": 666.777}}`),
			[]datatype.DataType{datatype.NewFloatType("Nested.FloatType", 666.777)},
		},
		{ //4
			"",
			[]byte(`{"Multy": 666.77, "Nested": {"FloatType": 666.999}}`),
			[]datatype.DataType{datatype.NewFloatType("Multy", 666.77), datatype.NewFloatType("Nested.FloatType", 666.999)},
		},
		{ //5
			"",
			[]byte(`{"Multy": 666.77, "Nested": {"FloatType": 666.999}}`),
			[]datatype.DataType{datatype.NewFloatType("Nested.FloatType", 666.999), datatype.NewFloatType("Multy", 666.77)},
		},
		{ //6
			"",
			[]byte(`{"FloatListType": []}`),
			[]datatype.DataType{datatype.NewFloatListType("FloatListType", []float64{})},
		},
		{ //7
			"",
			[]byte(`{"FloatListType": [0.1,1.2,2.3,3.4,666]}`),
			[]datatype.DataType{datatype.NewFloatListType("FloatListType", []float64{0.1, 1.2, 2.3, 3.4, 666})},
		},
		{ //8
			"",
			[]byte(`{"PauseNs": []}`),
			[]datatype.DataType{datatype.NewGCListType("PauseNs", []uint64{})},
		},
		{ //9
			"",
			[]byte(`{"PauseNs": [0,0,0,0,12481868021080215863,1481868005672005459,1481868012773129951,666000,11481937182104993300]}`),
			[]datatype.DataType{datatype.NewGCListType("PauseNs", []uint64{12481868021080215863, 1481868005672005459, 1481868012773129951, 666000, 11481937182104993300})},
		},
		{ //10
			"",
			[]byte(`{"TotalAlloc": 0}`),
			[]datatype.DataType{datatype.NewByteType("TotalAlloc", 0)},
		},
		{ //11
			"",
			[]byte(`{"TotalAlloc": 236478234}`),
			[]datatype.DataType{datatype.NewByteType("TotalAlloc", 236478234)},
		},
		{ //12
			"",
			[]byte(`{"PauseNs": [1481938891973801922,1481938893974355168,1481938895974915920,1481938897975467569,1481938899975919573,1481938901976464855,1481938903977051088,1481938905977636658,1481938907978221684,1481938909978619244,1481938911979100042,1481938913979740815,1481938915980232455,1481938917980671611,1481938919981183393,1481938921981827241,1481938923982308276,1481938925982865139,1481938927983327577,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}`),
			[]datatype.DataType{datatype.NewGCListType("PauseNs", []uint64{1481938891973801922, 1481938893974355168, 1481938895974915920, 1481938897975467569, 1481938899975919573, 1481938901976464855, 1481938903977051088, 1481938905977636658, 1481938907978221684, 1481938909978619244, 1481938911979100042, 1481938913979740815, 1481938915980232455, 1481938917980671611, 1481938919981183393, 1481938921981827241, 1481938923982308276, 1481938925982865139, 1481938927983327577})},
		},
		{ //13
			"",
			[]byte(`{"memstats": {"Alloc":2780496,"TotalAlloc": 236478234}}`),
			[]datatype.DataType{datatype.NewByteType("memstats.Alloc", 2780496), datatype.NewByteType("memstats.TotalAlloc", 236478234)},
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
