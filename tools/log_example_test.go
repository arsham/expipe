// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package tools_test

import (
	"github.com/arsham/expipe/tools"
)

// To get a logger with info level.
func ExampleGetLogger() {
	tools.GetLogger("info")
	// It's case insensitive.
	tools.GetLogger("INFO")
}
