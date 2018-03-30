// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype_test

import (
	"fmt"
	"testing"

	"github.com/arsham/expipe/datatype"
)

func TestFloatInSlice(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		niddle   float64
		haystack []float64
		result   bool
	}{
		{0, []float64{666.666, 666.777}, false},
		{666.666, []float64{666.666, 666.777}, true},
		{666.666, []float64{666.666, 666.666}, true},
		{666.666, []float64{666.777}, false},
		{666.666, []float64{}, false},
		{666.666, []float64{666.66}, false},
	}
	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			if ok := datatype.FloatInSlice(tc.niddle, tc.haystack); ok != tc.result {
				t.Errorf("FloatInSlice(tc.niddle, tc.haystack) = (%t); want (%t)", ok, tc.result)
			}
		})
	}
}

func TestUint64InSlice(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		niddle   uint64
		haystack []uint64
		result   bool
	}{
		{0, []uint64{666, 6662}, false},
		{666, []uint64{666}, true},
		{666, []uint64{666, 666}, true},
		{6666666666, []uint64{6666666666}, true},
		{666, []uint64{}, false},
	}
	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			if ok := datatype.Uint64InSlice(tc.niddle, tc.haystack); ok != tc.result {
				t.Errorf("Uint64InSlice(tc.niddle, tc.haystack) = (%t); want (%t)", ok, tc.result)
			}
		})
	}
}
