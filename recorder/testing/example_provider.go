// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"time"

	"github.com/arsham/expvastic/internal"
)

// GetRecorder provides a SimpleRecorder for using in the example.
func GetRecorder(ctx context.Context, url string) *Recorder {
	log := internal.DiscardLogger()
	rec, err := New(ctx, log, "reader_example", url, "intexName", time.Second, 5)
	if err != nil {
		panic(err)
	}
	return rec
}
