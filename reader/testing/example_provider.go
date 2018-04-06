// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"time"

	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/tools"
)

// GetReader provides a MockReader for using in the example.
func GetReader(url string) *Reader {
	log := tools.DiscardLogger()
	red, err := New(
		reader.WithLogger(log),
		reader.WithEndpoint(url),
		reader.WithName("reader_example"),
		reader.WithTypeName("reader_example"),
		reader.WithInterval(10*time.Millisecond),
		reader.WithTimeout(time.Second),
	)
	if err != nil {
		panic(err)
	}
	return red
}
