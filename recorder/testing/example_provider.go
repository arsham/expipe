// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/recorder"
)

// GetRecorder provides a SimpleRecorder for using in the example.
func GetRecorder(url string) *Recorder {
	log := internal.DiscardLogger()
	red, err := New(
		recorder.SetLogger(log),
		recorder.SetEndpoint(url),
		recorder.SetName("recorder_example"),
		recorder.SetIndexName("recorder_example"),
		recorder.SetTimeout(time.Second),
		recorder.SetBackoff(5),
	)
	if err != nil {
		panic(err)
	}
	return red
}
