// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package lib

// StringInSlice returns true if niddle is in the haystack
func StringInSlice(niddle string, haystack []string) bool {
    for _, b := range haystack {
        if b == niddle {
            return true
        }
    }
    return false
}
