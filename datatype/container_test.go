// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/arsham/expipe/datatype"
	"github.com/pkg/errors"
)

var errExample = errors.New("DataType Error")

type badDataType struct{}

func (badDataType) Read([]byte) (int, error)     { return 0, errExample }
func (badDataType) Equal(datatype.DataType) bool { return true }
func (badDataType) Reset()                       {}

func inArray(a datatype.DataType, b []datatype.DataType) bool {
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

func TestList(t *testing.T) {
	l := []datatype.DataType{&datatype.FloatType{}}
	c := datatype.New(l)
	if !reflect.DeepEqual(c.List(), l) {
		t.Errorf("reflect.DeepEqual(): (%v) and (%v) lists are not equal", c.List(), l)
	}
}

func TestJobResultDataTypes(t *testing.T) {
	mapper := datatype.DefaultMapper()
	tcs := []struct {
		name  string
		input []byte
		err   error
	}{
		{
			"missing leading {",
			[]byte(`"memstats": {"PauseNs":[666,777]}}`),
			errExample,
		},
		{
			"missing ending }",
			[]byte(`{"memstats": {"PauseNs":[666,777]}`),
			errExample,
		},
		{
			"simple string",
			[]byte(`"memstats PauseNs 666 777"`),
			errExample,
		},
		{
			"string instead of float",
			[]byte(`{"memstats": {"PauseNs":["666"]}}`),
			datatype.ErrUnidentifiedJason,
		},
		{
			"float instead of int",
			[]byte(`{"memstats": {"TotalAlloc":[666.5]}}`),
			datatype.ErrUnidentifiedJason,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := datatype.JobResultDataTypes(tc.input, mapper)
			if tc.err == errExample {
				if err == nil {
					t.Error("err = (nil); want (error)")
				}
			} else {
				if err != tc.err {
					t.Errorf("err = (%v); got (%v)", err, tc.err)
				}
			}
		})
	}

	tcs2 := []struct {
		name  string
		input []byte
		exp   []datatype.DataType
	}{
		{
			"one value",
			[]byte(`{"memstats": {"TotalAlloc":666}}`),
			[]datatype.DataType{
				datatype.NewMegaByteType("memstats.TotalAlloc", 666),
				datatype.NewMegaByteType("memstats.TotalAlloc", 666),
			},
		},
		{
			"multiple values",
			[]byte(`{"memstats": {"TotalAlloc":666, "HeapIdle":777}}`),
			[]datatype.DataType{
				datatype.NewMegaByteType("memstats.TotalAlloc", 666),
				datatype.NewMegaByteType("memstats.HeapIdle", 777),
				datatype.NewMegaByteType("memstats.TotalAlloc", 666),
				datatype.NewMegaByteType("memstats.HeapIdle", 777),
			},
		},
	}

	for _, tc := range tcs2 {
		t.Run(tc.name, func(t *testing.T) {
			c, err := datatype.JobResultDataTypes(tc.input, mapper)
			l := datatype.New(nil)
			l.Add(tc.exp...)
			if errors.Cause(err) != nil {
				t.Errorf("err = (%#v); want (nil)", err)
			}
			for _, i := range l.List() {
				if !inArray(i, c.List()) {
					t.Errorf("inArray(i, c.List()): want (%v) be in (%v)", i, l.List())
				}
			}
		})
	}
}

func TestInArray(t *testing.T) {
	a := datatype.NewStringType("key", "value")
	aa := datatype.NewStringType("key", "value1")
	b := datatype.NewFloatType("key", 6.66)

	tcs := []struct {
		name  string
		left  datatype.DataType
		right []datatype.DataType
	}{
		{"a in nothing", a, []datatype.DataType{}},
		{"a in aa", a, []datatype.DataType{aa}},
		{"a in b", a, []datatype.DataType{b}},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if inArray(tc.left, tc.right) {
				t.Error("inArray(tc.left, tc.right) = true; want (false)")
			}
		})
	}
	if !inArray(a, []datatype.DataType{a, aa}) {
		t.Error("a, []datatype.DataType{a, aa} = false; want (true)")
	}
	if !inArray(a, []datatype.DataType{a, b}) {
		t.Error("inArray(a, []datatype.DataType{a, b}) = false; want (true)")
	}
}
