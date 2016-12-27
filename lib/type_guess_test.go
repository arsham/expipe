// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package lib

import (
	"fmt"
	"testing"
)

func TestIsGCType(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		input    string
		expected bool
	}{
		{"PauseEnd", true},
		{"PauseNs", true},
		{"memstats.PauseEnd", true},
		{"memstats.PauseNs", true},
		{"pauseend", false},
		{"pausens", false},
		{"memstats.Pauseend", false},
		{"memstats.Pausens", false},
		{"", false},
	}
	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			if res := IsGCType(tc.input); res != tc.expected {
				t.Errorf("want (%t), got (%t)", tc.expected, res)
			}
		})
	}
}

func TestIsMBType(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		input    string
		expected bool
	}{

		{"Alloc", true},
		{"alloc", false},
		{"TotalAlloc", true},
		{"totalalloc", false},
		{"Sys", true},
		{"sys", false},
		{"HeapAlloc", true},
		{"heapalloc", false},
		{"HeapSys", true},
		{"heapsys", false},
		{"HeapIdle", true},
		{"heapidle", false},
		{"HeapInuse", true},
		{"heapinuse", false},
		{"HeapReleased", true},
		{"heapreleased", false},
		{"StackInuse", true},
		{"stackinuse", false},
		{"memstats.Alloc", true},
		{"memstats.alloc", false},
		{"memstats.TotalAlloc", true},
		{"memstats.totalalloc", false},
		{"memstats.Sys", true},
		{"memstats.sys", false},
		{"memstats.HeapAlloc", true},
		{"memstats.heapalloc", false},
		{"memstats.HeapSys", true},
		{"memstats.heapsys", false},
		{"memstats.HeapIdle", true},
		{"memstats.heapidle", false},
		{"memstats.HeapInuse", true},
		{"memstats.heapinuse", false},
		{"memstats.HeapReleased", true},
		{"memstats.heapreleased", false},
		{"memstats.StackInuse", true},
		{"memstats.stackinuse", false},
		{"", false},
	}
	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			if res := IsMBType(tc.input); res != tc.expected {
				t.Errorf("want (%t), got (%t)", tc.expected, res)
			}
		})
	}
}
