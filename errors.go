// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import "fmt"

var (
    // ErrEmptyRecName is for when the recorder's name is an empty string.
    ErrEmptyRecName = fmt.Errorf("recorder name empty")

    // ErrDupRecName is for when there are two recorders with the same name.
    ErrDupRecName = fmt.Errorf("recorder name empty")
)
