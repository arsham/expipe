// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package lib

import (
	"fmt"
	"testing"
)

// StringInSlice returns true if niddle is in the haystack
func TestStringInSlice(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		niddle   string
		haystack []string
		result   bool
	}{
		{"aaa", []string{"aaa", "bbb"}, true},
		{"aaa", []string{"aaa", "aaa"}, true},
		{"aaa", []string{"bbb"}, false},
		{"aaa", []string{}, false},
		{"aaa", []string{"aaaa"}, false},
		{"aaa", []string{"AAA"}, false},
	}
	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if ok := StringInSlice(tc.niddle, tc.haystack); ok != tc.result {
				t.Errorf("want (%t), got (%t)", tc.result, ok)
			}
		})
	}
}

// StringInSlice returns true if niddle is in the haystack
func TestStringInMapKeys(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		niddle   string
		haystack map[string]string
		result   bool
	}{
		{"aaa", map[string]string{"aaa": "a"}, true},
		{"aaa", map[string]string{"aaa": "a", "bbbb": "a"}, true},
		{"aaa", map[string]string{"bbb": "a"}, false},
		{"aaa", map[string]string{"aaaa": "a"}, false},
		{"aaa", map[string]string{"AAA": "a"}, false},
	}
	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if ok := StringInMapKeys(tc.niddle, tc.haystack); ok != tc.result {
				t.Errorf("want (%t), got (%t)", tc.result, ok)
			}
		})
	}
}
