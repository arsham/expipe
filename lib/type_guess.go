// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package lib

var mbtypes = []string{
    "Alloc",
    "TotalAlloc",
    "Sys",
    "HeapAlloc",
    "HeapSys",
    "HeapIdle",
    "HeapInuse",
    "HeapReleased",
    "StackInuse",
}

// IsGCType returns true if the key corresponds to one
func IsGCType(key string) bool {
    return StringInSlice(key, []string{"PauseEnd", "PauseNs"})
}

// IsMBType returns true if key's value is a large byte value
func IsMBType(key string) bool {
    return StringInSlice(key, mbtypes)
}
