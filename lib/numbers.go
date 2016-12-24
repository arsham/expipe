// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package lib

// FloatInSlice returns true if niddle is in the haystack
func FloatInSlice(niddle float64, haystack []float64) bool {
    for _, b := range haystack {
        if b == niddle {
            return true
        }
    }
    return false
}

// Uint64InSlice returns true if niddle is in the haystack
func Uint64InSlice(niddle uint64, haystack []uint64) bool {
    for _, b := range haystack {
        if b == niddle {
            return true
        }
    }
    return false
}
