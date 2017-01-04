// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import "fmt"

// ErrDuplicateRecorderName is for when there are two recorders with the same name.
var ErrDuplicateRecorderName = fmt.Errorf("recorder name cannot be reused")

// ErrPing is the error when one of readers/recorder has a ping error
type ErrPing struct {
	Name string
	Err  error
}

// Ping defines the behaviour of the error
func (ErrPing) Ping() {}
func (e ErrPing) Error() string {
	return fmt.Sprintf("pinging (%s) error: %s", e.Name, e.Err)
}
