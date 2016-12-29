// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype_test

import (
	"fmt"
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
			results := contaner.String(now)
			if results != tc.expected {
				t.Errorf("want (%s) got (%s)", tc.expected, results)
			}
		})
	}
}
