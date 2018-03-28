// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/arsham/expipe/datatype"
	"github.com/pkg/errors"
)

var errExample = errors.New("DataType Error")

type badDataType struct{}

func (badDataType) Write(io.Writer) (int, error) { return 0, errExample }
func (badDataType) Equal(datatype.DataType) bool { return true }

func inArray(a datatype.DataType, b []datatype.DataType) bool {
	ap := new(bytes.Buffer)
	a.Write(ap)
	for i := range b {
		bp := new(bytes.Buffer)
		b[i].Write(bp)
		if reflect.DeepEqual(ap.Bytes(), bp.Bytes()) {
			return true
		}
	}
	return false
}

func TestList(t *testing.T) {
	l := []datatype.DataType{datatype.FloatType{}}
	c := datatype.New(l)
	if !reflect.DeepEqual(c.List(), l) {
		t.Errorf("lists are not equal: (%v) and (%v)", c.List(), l)
	}
}

func TestJobResultDataTypes(t *testing.T) {
	mapper := datatype.DefaultMapper()
	tcs := []struct {
		name  string
		input []byte
		err   error
	}{
		{"missing leading {", []byte(`"memstats": {"PauseNs":[666,777]}}`), errExample},
		{"missing ending }", []byte(`{"memstats": {"PauseNs":[666,777]}`), errExample},
		{"simple string", []byte(`"memstats PauseNs 666 777"`), errExample},
		{"string instead of float", []byte(`{"memstats": {"PauseNs":["666"]}}`), datatype.ErrUnidentifiedJason},
		{"float instead of int", []byte(`{"memstats": {"TotalAlloc":[666.5]}}`), datatype.ErrUnidentifiedJason},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := datatype.JobResultDataTypes(tc.input, mapper)
			if tc.err == errExample {
				if err == nil {
					t.Error("want (error), got (nil)")
				}
			} else {
				if err != tc.err {
					t.Errorf("want (%v), got (%v)", tc.err, err)
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
				&datatype.MegaByteType{Key: "memstats.TotalAlloc", Value: 666},
				&datatype.MegaByteType{Key: "memstats.TotalAlloc", Value: 666},
			},
		},
		{
			"multiple values",
			[]byte(`{"memstats": {"TotalAlloc":666, "HeapIdle":777}}`),
			[]datatype.DataType{
				&datatype.MegaByteType{Key: "memstats.TotalAlloc", Value: 666},
				&datatype.MegaByteType{Key: "memstats.HeapIdle", Value: 777},
				&datatype.MegaByteType{Key: "memstats.TotalAlloc", Value: 666},
				&datatype.MegaByteType{Key: "memstats.HeapIdle", Value: 777},
			},
		},
	}

	for _, tc := range tcs2 {
		t.Run(tc.name, func(t *testing.T) {
			c, err := datatype.JobResultDataTypes(tc.input, mapper)
			l := datatype.New(nil)
			l.Add(tc.exp...)
			if errors.Cause(err) != nil {
				t.Errorf("want (nil), got (%#v)", err)
			}
			for _, i := range l.List() {
				if !inArray(i, c.List()) {
					t.Errorf("want (%v) be in (%v)", i, l.List())
				}
			}
		})
	}
}
