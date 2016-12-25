// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
	"fmt"
	"testing"
	"time"
)

func TestGetQueryString(t *testing.T) {
	now := time.Now()
	tStr := fmt.Sprintf(`"@timestamp":"%s"`, now.Format("2006-01-02T15:04:05.999999-07:00"))

	testCase := []struct {
		input    []DataType
		expected string
	}{
		{
			[]DataType{},
			fmt.Sprintf("{%s}", tStr),
		},
		{
			[]DataType{&FloatType{"test", 3.4}},
			fmt.Sprintf(`{%s,"test":%f}`, tStr, 3.4),
		},
		{
			[]DataType{&StringType{"test", "3a"}, &FloatType{"test2", 2.2}},
			fmt.Sprintf(`{%s,"test":"%s","test2":%f}`, tStr, "3a", 2.2),
		},
	}

	for i, tc := range testCase {
		name := fmt.Sprintf("case %d", i)
		t.Run(name, func(t *testing.T) {
			contaner := NewContainer(tc.input)
			results := contaner.String(now)
			if results != tc.expected {
				t.Errorf("want (%s) got (%s)", tc.expected, results)
			}
		})
	}
}
