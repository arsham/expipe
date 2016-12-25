// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import "bytes"

func defaultMappings() *bytes.Buffer {
	return bytes.NewBuffer([]byte(`
gc_types:
    PauseEnd
    PauseNs
    memstats.PauseEnd
    memstats.PauseNs
memory_bytes:
    Alloc: mb
    TotalAlloc: mb
    Sys: mb
    HeapAlloc: mb
    HeapSys: mb
    HeapIdle: mb
    HeapInuse: mb
    HeapReleased: mb
    StackInuse: mb
    memstats.Alloc: mb
    memstats.TotalAlloc: mb
    memstats.Sys: mb
    memstats.HeapAlloc: mb
    memstats.HeapSys: mb
    memstats.HeapIdle: mb
    memstats.HeapInuse: mb
    memstats.HeapReleased: mb
    memstats.StackInuse: mb
`))
}
