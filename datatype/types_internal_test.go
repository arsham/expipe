// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/antonholmquist/jason"
)

type caseType struct {
	prefix   string
	value    []byte
	expected []DataType
}

func inArray(a DataType, b []DataType) bool {
	ap := new(bytes.Buffer)
	ap.ReadFrom(a)
	for i := range b {
		bp := new(bytes.Buffer)
		bp.ReadFrom(b[i])
		if strings.Contains(ap.String(), bp.String()) {
			return true
		}
	}
	return false
}

func isIn(a, b []DataType) bool {
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

func testCase() []caseType {
	return []caseType{
		{ //0
			"",
			[]byte(`{"FloatType": 123.4}`),
			[]DataType{NewFloatType("FloatType", 123.4)},
		},
		{ //1
			"",
			[]byte(`{"StringType": "Random: 666"}`),
			[]DataType{NewStringType("StringType", "Random: 666")},
		},
		{ //2
			"aaa.",
			[]byte(`{"Prefixed": 666.777}`),
			[]DataType{NewFloatType("aaa.Prefixed", 666.777)},
		},
		{ //3
			"",
			[]byte(`{"Nested": {"FloatType": 666.777}}`),
			[]DataType{NewFloatType("Nested.FloatType", 666.777)},
		},
		{ //4
			"",
			[]byte(`{"Multy": 666.77, "Nested": {"FloatType": 666.999}}`),
			[]DataType{NewFloatType("Multy", 666.77), NewFloatType("Nested.FloatType", 666.999)},
		},
		{ //5
			"",
			[]byte(`{"Multy": 666.77, "Nested": {"FloatType": 666.999}}`),
			[]DataType{NewFloatType("Nested.FloatType", 666.999), NewFloatType("Multy", 666.77)},
		},
		{ //6
			"",
			[]byte(`{"FloatListType": []}`),
			[]DataType{NewFloatListType("FloatListType", []float64{})},
		},
		{ //7
			"",
			[]byte(`{"FloatListType": [0.1,1.2,2.3,3.4,666]}`),
			[]DataType{NewFloatListType("FloatListType", []float64{0.1, 1.2, 2.3, 3.4, 666})},
		},
		{ //8
			"",
			[]byte(`{"PauseNs": []}`),
			[]DataType{NewGCListType("PauseNs", []uint64{})},
		},
		{ //9
			"",
			[]byte(`{"PauseNs": [0,0,0,0,12481868021080215863,1481868005672005459,1481868012773129951,666000,11481937182104993300]}`),
			[]DataType{NewGCListType("PauseNs", []uint64{12481868021080215863, 1481868005672005459, 1481868012773129951, 666000, 11481937182104993300})},
		},
		{ //10
			"",
			[]byte(`{"TotalAlloc": 0}`),
			[]DataType{NewByteType("TotalAlloc", 0)},
		},
		{ //11
			"",
			[]byte(`{"TotalAlloc": 236478234}`),
			[]DataType{NewByteType("TotalAlloc", 236478234)},
		},
		{ //12
			"",
			[]byte(`{"PauseNs": [1481938891973801922,1481938893974355168,1481938895974915920,1481938897975467569,1481938899975919573,1481938901976464855,1481938903977051088,1481938905977636658,1481938907978221684,1481938909978619244,1481938911979100042,1481938913979740815,1481938915980232455,1481938917980671611,1481938919981183393,1481938921981827241,1481938923982308276,1481938925982865139,1481938927983327577,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}`),
			[]DataType{NewGCListType("PauseNs", []uint64{1481938891973801922, 1481938893974355168, 1481938895974915920, 1481938897975467569, 1481938899975919573, 1481938901976464855, 1481938903977051088, 1481938905977636658, 1481938907978221684, 1481938909978619244, 1481938911979100042, 1481938913979740815, 1481938915980232455, 1481938917980671611, 1481938919981183393, 1481938921981827241, 1481938923982308276, 1481938925982865139, 1481938927983327577})},
		},
		{ //13
			"",
			[]byte(`{"memstats": {"Alloc":2780496,"TotalAlloc": 236478234}}`),
			[]DataType{NewByteType("memstats.Alloc", 2780496), NewByteType("memstats.TotalAlloc", 236478234)},
		},
	}
}

func TestGetJasonValues(t *testing.T) {
	t.Parallel()
	mapper := &MapConvertMock{}
	for i, tc := range testCase() {
		name := fmt.Sprintf("case %d", i)
		t.Run(name, func(t *testing.T) {
			var payload []DataType
			obj, _ := jason.NewObjectFromBytes(tc.value)
			payload = append(payload, mapper.Values(tc.prefix, obj.Map())...)

			if len(payload) == 0 {
				t.Errorf("len(payload) = (%d); want (%d)", len(payload), len(tc.expected))
				return
			}
			results := New(payload)
			if !isIn(results.List(), tc.expected) {
				t.Errorf("isIn(results.List(), tc.expected): results.List() = (%v); want (%v)", results.List(), tc.expected)
			}
		})
	}
}

func TestGetJasonValuesAddToContainer(t *testing.T) {
	t.Parallel()
	mapper := &MapConvertMock{}
	for i, tc := range testCase() {
		name := fmt.Sprintf("case %d", i)
		var container Container
		t.Run(name, func(t *testing.T) {
			obj, _ := jason.NewObjectFromBytes(tc.value)
			for _, value := range mapper.Values(tc.prefix, obj.Map()) {
				container.Add(value)
			}
			if !isIn(container.List(), tc.expected) {
				t.Errorf("isIn(container.List(), tc.expected): container.List() = (%#v); want (%#v)", container.List(), tc.expected)
			}
		})
	}
}

func TestFromReader(t *testing.T) {
	t.Parallel()
	mapper := &MapConvertMock{}
	for i, tc := range testCase() {
		if tc.prefix != "" {
			continue
		}
		name := fmt.Sprintf("case %d", i)
		t.Run(name, func(t *testing.T) {
			results, err := JobResultDataTypes(tc.value, mapper)
			if err != nil {
				t.Errorf("err = (%s); want (nil)", err)
			}
			if !isIn(results.List(), tc.expected) {
				t.Errorf("isIn(results.List(), tc.expected): results.List() = (%s); want (%s)", results.List(), tc.expected)
			}
		})
	}
}

func TestJobResultDataTypesErrors(t *testing.T) {
	t.Parallel()
	mapper := &MapConvertMock{}

	value := []byte(`{"Alloc": "sdsds"}`)
	results, err := JobResultDataTypes(value, mapper)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if results != nil {
		t.Errorf("results = (%v); want (nil)", results)
	}

	value = []byte(`{"Alloc": "sdsds}`)
	results, err = JobResultDataTypes(value, mapper)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if results != nil {
		t.Errorf("results = (%v); want (nil)", results)
	}
}

func TestInArray(t *testing.T) {
	a := NewStringType("key", "value")
	aa := NewStringType("key", "value1")
	b := NewFloatType("key", 6.66)

	tcs := []struct {
		name  string
		left  DataType
		right []DataType
	}{
		{"a in nothing", a, []DataType{}},
		{"a in aa", a, []DataType{aa}},
		{"a in b", a, []DataType{b}},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if inArray(tc.left, tc.right) {
				t.Error("wrong!")
			}
		})
	}
	if !inArray(a, []DataType{a, aa}) {
		t.Error("inArray(a, []DataType{a, aa}) = (false); want (true)")
	}
	if !inArray(a, []DataType{a, b}) {
		t.Error("inArray(a, []DataType{a, b}) = (false); want (true)")
	}
}

func TestIsIn(t *testing.T) {
	a := NewStringType("key", "value")
	aa := NewStringType("key", "value")
	b := NewFloatType("key", 6.66)
	c := NewFloatType("key2", 6.66)

	tcs := []struct {
		name   string
		left   []DataType
		right  []DataType
		result bool
	}{
		{"a in nothing", []DataType{a}, []DataType{}, false},
		{"a in aa", []DataType{a}, []DataType{aa}, true},
		{"a in b", []DataType{a}, []DataType{b}, false},
		{"ab in ba", []DataType{a, b}, []DataType{b, a}, true},
		{"abc in bca", []DataType{a, b, c}, []DataType{b, c, a}, true},
		{"ab in bca", []DataType{a, b}, []DataType{b, c, a}, false},
		{"bca in ab", []DataType{b, c, a}, []DataType{a, b}, false},
		{"abc in ab", []DataType{a, b, c}, []DataType{a, b}, false},
		{"aab in baa", []DataType{a, a, b}, []DataType{b, a, a}, true},
		{"aab in aba", []DataType{a, a, b}, []DataType{b, a, a}, true},
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
