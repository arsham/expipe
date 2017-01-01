// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/antonholmquist/jason"
)

type caseType struct {
	prefix   string
	value    []byte
	expected []DataType
}

func inArray(a DataType, b []DataType) bool {
	for i := range b {
		if reflect.DeepEqual(a.Bytes(), b[i].Bytes()) {
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
			[]DataType{&FloatType{"FloatType", 123.4}},
		},
		{ //1
			"",
			[]byte(`{"StringType": "Random: 666"}`),
			[]DataType{&StringType{"StringType", "Random: 666"}},
		},
		{ //2
			"aaa.",
			[]byte(`{"Prefixed": 666.777}`),
			[]DataType{&FloatType{"aaa.Prefixed", 666.777}},
		},
		{ //3
			"",
			[]byte(`{"Nested": {"FloatType": 666.777}}`),
			[]DataType{&FloatType{"Nested.FloatType", 666.777}},
		},
		{ //4
			"",
			[]byte(`{"Multy": 666.77, "Nested": {"FloatType": 666.999}}`),
			[]DataType{&FloatType{"Multy", 666.77}, &FloatType{"Nested.FloatType", 666.999}},
		},
		{ //5
			"",
			[]byte(`{"Multy": 666.77, "Nested": {"FloatType": 666.999}}`),
			[]DataType{&FloatType{"Nested.FloatType", 666.999}, &FloatType{"Multy", 666.77}},
		},
		{ //6
			"",
			[]byte(`{"FloatListType": []}`),
			[]DataType{&FloatListType{"FloatListType", []float64{}}},
		},
		{ //7
			"",
			[]byte(`{"FloatListType": [0.1,1.2,2.3,3.4,666]}`),
			[]DataType{&FloatListType{"FloatListType", []float64{0.1, 1.2, 2.3, 3.4, 666}}},
		},
		{ //8
			"",
			[]byte(`{"PauseNs": []}`),
			[]DataType{&GCListType{"PauseNs", []uint64{}}},
		},
		{ //9
			"",
			[]byte(`{"PauseNs": [0,0,0,0,12481868021080215863,1481868005672005459,1481868012773129951,666000,11481937182104993300]}`),
			[]DataType{&GCListType{"PauseNs", []uint64{12481868021080215863, 1481868005672005459, 1481868012773129951, 666000, 11481937182104993300}}},
		},
		{ //10
			"",
			[]byte(`{"TotalAlloc": 0}`),
			[]DataType{&ByteType{"TotalAlloc", 0}},
		},
		{ //11
			"",
			[]byte(`{"TotalAlloc": 236478234}`),
			[]DataType{&ByteType{"TotalAlloc", 236478234}},
		},
		{ //12
			"",
			[]byte(`{"PauseNs": [1481938891973801922,1481938893974355168,1481938895974915920,1481938897975467569,1481938899975919573,1481938901976464855,1481938903977051088,1481938905977636658,1481938907978221684,1481938909978619244,1481938911979100042,1481938913979740815,1481938915980232455,1481938917980671611,1481938919981183393,1481938921981827241,1481938923982308276,1481938925982865139,1481938927983327577,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}`),
			[]DataType{&GCListType{"PauseNs", []uint64{1481938891973801922, 1481938893974355168, 1481938895974915920, 1481938897975467569, 1481938899975919573, 1481938901976464855, 1481938903977051088, 1481938905977636658, 1481938907978221684, 1481938909978619244, 1481938911979100042, 1481938913979740815, 1481938915980232455, 1481938917980671611, 1481938919981183393, 1481938921981827241, 1481938923982308276, 1481938925982865139, 1481938927983327577}}},
		},
		{ //13
			"",
			[]byte(`{"memstats": {"Alloc":2780496,"TotalAlloc": 236478234}}`),
			[]DataType{&ByteType{"memstats.Alloc", 2780496}, &ByteType{"memstats.TotalAlloc", 236478234}},
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
				t.Errorf("want (%d), got (%d)", len(tc.expected), len(payload))
				return
			}
			results := NewContainer(payload)
			if !isIn(results.List(), tc.expected) {
				t.Errorf("expected (%v), got (%v)", tc.expected, results.List())
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
				t.Errorf("expected (%v), got (%v)", tc.expected, container.List())
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

			results := JobResultDataTypes(tc.value, mapper)
			if results.Error() != nil {
				t.Errorf("expected no errors, got (%s)", results.Error())
			}
			if !isIn(results.List(), tc.expected) {
				t.Errorf("want (%s) got (%s)", tc.expected, results.List())
			}
		})
	}
}

func TestJobResultDataTypesErrors(t *testing.T) {
	t.Parallel()
	mapper := &MapConvertMock{}

	value := []byte(`{"Alloc": "sdsds"}`)
	results := JobResultDataTypes(value, mapper)
	if results.Error() == nil {
		t.Error("expected error, got nothing")
	}
	if results.Len() != 0 {
		t.Errorf("expected empty results, got (%s)", results.List())
	}

	value = []byte(`{"Alloc": "sdsds}`)
	results = JobResultDataTypes(value, mapper)
	if results.Error() == nil {
		t.Error("expected error, got nothing")
	}
	if results.Len() != 0 {
		t.Errorf("expected empty results, got (%s)", results.List())
	}
}
