// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/reader"
)

// GetReader provides a SimpleReader for using in the example.
func GetReader(url string) *Reader {
	log := internal.DiscardLogger()
	red, err := New(
		reader.SetLogger(log),
		reader.SetEndpoint(url),
		reader.SetName("reader_example"),
		reader.SetTypeName("reader_example"),
		reader.SetInterval(10*time.Millisecond),
		reader.SetTimeout(time.Second),
		reader.SetBackoff(10),
	)
	if err != nil {
		panic(err)
	}
	return red
}
