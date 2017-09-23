// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expipe

import (
	"fmt"
	"strings"
)

var (
	// ErrNoReader is returned when no reader has been provided
	ErrNoReader = fmt.Errorf("no reader provided")

	// ErrNoLogger is returned when no logger has been provided
	ErrNoLogger = fmt.Errorf("no logger provided")

	// ErrNoCtx is returned when no ctx has been provided
	ErrNoCtx = fmt.Errorf("no ctx provided")
)

// ErrPing is the error when one of readers/recorder has a ping error
type ErrPing map[string]error

func (e ErrPing) Error() string {
	var msgs []string
	for name, err := range e {
		msgs = append(msgs, name+":"+err.Error())
	}
	return fmt.Sprintf("pinging error: %s", strings.Join(msgs, "\n"))
}
