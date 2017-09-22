// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"time"

	"github.com/arsham/expvastic/internal"
)

// GetReader provides a SimpleReader for using in the example.
func GetReader(url string) *Reader {
	log := internal.DiscardLogger()
	red, err := New(log, url, "reader_example", "reader_example", 10*time.Millisecond, 10*time.Millisecond, 10)
	if err != nil {
		panic(err)
	}
	return red
}
