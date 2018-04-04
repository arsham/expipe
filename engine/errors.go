// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"strings"

	"github.com/arsham/expipe/tools/token"
)

// Errors returning from Engine operations.
var (
	ErrNoReader   = fmt.Errorf("no reader provided")
	ErrNoRecorder = fmt.Errorf("no recorder provided")
	ErrNoLogger   = fmt.Errorf("no logger provided")
	ErrNoCtx      = fmt.Errorf("no ctx provided")
)

// PingError is the error when one of readers/recorder has a ping error.
type PingError map[string]error

func (e PingError) Error() string {
	var msgs []string
	for name, err := range e {
		msgs = append(msgs, name+":"+err.Error())
	}
	return fmt.Sprintf("pinging error: %s", strings.Join(msgs, "\n"))
}

// JobError caries an error around in Engine operations.
type JobError struct {
	Name string // Name of the operator; reader, recorder.
	ID   token.ID
	Err  error
}

func (e JobError) Error() string {
	return fmt.Sprintf("%s - [ID %s]: %s", e.Name, e.ID.String(), e.Err.Error())
}
