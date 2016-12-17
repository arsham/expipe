// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
    "encoding/json"
    "fmt"
    "io"
    "strings"
    "testing"
)

func TestConvertToActual(t *testing.T) {
    testCase := []struct {
        prefix   string
        value    io.Reader
        expected []DataType
    }{
        { //0
            "",
            strings.NewReader(`{"FloatType": 123.4}`),
            []DataType{&FloatType{"FloatType", 123.4}},
        },
        { //1
            "",
            strings.NewReader(`{"StringType": "Random: 666"}`),
            []DataType{&StringType{"StringType", "Random: 666"}},
        },
        { //2
            "aaa.",
            strings.NewReader(`{"Prefixed": 666.777}`),
            []DataType{&FloatType{"aaa.Prefixed", 666.777}},
        },
        { //3
            "",
            strings.NewReader(`{"Nested": {"FloatType": 666.777}}`),
            []DataType{&FloatType{"Nested.FloatType", 666.777}},
        },
        { //4
            "",
            strings.NewReader(`{"Multy": 666.77, "Nested": {"FloatType": 666.999}}`),
            []DataType{&FloatType{"Multy", 666.77}, &FloatType{"Nested.FloatType", 666.999}},
        },
        { //5
            "",
            strings.NewReader(`{"FloatListType": []}`),
            []DataType{&FloatListType{"FloatListType", []float64{}}},
        },
        { //6
            "",
            strings.NewReader(`{"FloatListType": [0.1,1.2,2.3,3.4,666]}`),
            []DataType{&FloatListType{"FloatListType", []float64{0.1, 1.2, 2.3, 3.4, 666}}},
        },
        { //7
            "",
            strings.NewReader(`{"PauseNs": []}`),
            []DataType{&GCListType{"PauseNs", []uint64{}}},
        },
        { //8
            "",
            strings.NewReader(`{"PauseNs": [0,0,0,0,12481868021080215863,1481868005672005459,1481868012773129951,666000,11481937182104993300]}`),
            []DataType{&GCListType{"PauseNs", []uint64{12481868021080215863, 1481868005672005459, 1481868012773129951, 666000, 11481937182104993300}}},
        },
        { //9
            "",
            strings.NewReader(`{"TotalAlloc": 0}`),
            []DataType{&ByteType{"TotalAlloc", 0}},
        },
        { //10
            "",
            strings.NewReader(`{"TotalAlloc": 236478234}`),
            []DataType{&ByteType{"TotalAlloc", 236478234}},
        },
        { //11
            "",
            strings.NewReader(`{"PauseNs": [1481938891973801922,1481938893974355168,1481938895974915920,1481938897975467569,1481938899975919573,1481938901976464855,1481938903977051088,1481938905977636658,1481938907978221684,1481938909978619244,1481938911979100042,1481938913979740815,1481938915980232455,1481938917980671611,1481938919981183393,1481938921981827241,1481938923982308276,1481938925982865139,1481938927983327577,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}`),
            []DataType{&GCListType{"PauseNs", []uint64{1481938891973801922, 1481938893974355168, 1481938895974915920, 1481938897975467569, 1481938899975919573, 1481938901976464855, 1481938903977051088, 1481938905977636658, 1481938907978221684, 1481938909978619244, 1481938911979100042, 1481938913979740815, 1481938915980232455, 1481938917980671611, 1481938919981183393, 1481938921981827241, 1481938923982308276, 1481938925982865139, 1481938927983327577}}},
        },
    }
    isIn := func(a, b []DataType) bool {
        if len(a) != len(b) {
            return false
        }
        for i := range a {
            if a[i].String() != b[i].String() {
                return false
            }
        }
        return true
    }
    for i, tc := range testCase {
        name := fmt.Sprintf("case %d", i)
        t.Run(name, func(t *testing.T) {
            var mar map[string]interface{}
            json.NewDecoder(tc.value).Decode(&mar)
            result := convertToActual(tc.prefix, mar)
            if !isIn(result, tc.expected) {
                t.Errorf("want (%s) got (%s)", tc.expected, result)
            }
        })
    }
}
