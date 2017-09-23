// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config

import "fmt"

// EmptyConfigErr is an error when the config file is empty
var EmptyConfigErr = &StructureErr{"", "empty configuration file", nil}

// StructureErr is an error on reading the configuration file.
type StructureErr struct {
	Section string // The section that error happened
	Reason  string // The reason behind the error
	Err     error  // Err is the error that occurred during the operation.
}

const (
	nilStr = "<nil>"
)

// Error returns "<nil>" if the error is nil.
func (e *StructureErr) Error() string {
	if e == nil {
		return nilStr
	}

	s := e.Section
	if e.Reason != "" {
		s += " " + e.Reason
	}

	if e.Err != nil {
		s += ": " + e.Err.Error()
	}
	return s
}

type ErrNotSpecified StructureErr

func NewErrNotSpecified(section, reason string, err error) *ErrNotSpecified {
	return &ErrNotSpecified{section, reason, err}
}

func (e *ErrNotSpecified) Error() string {
	if e == nil {
		return nilStr
	}

	s := e.Section
	if e.Reason != "" {
		s += " " + e.Reason
	}
	s += " not specified"

	if e.Err != nil {
		s += ": " + e.Err.Error()
	}
	return s
}

// ErrRouters represents an error when routes are not configured correctly.
// The section on this error is the subsection of the route.
type ErrRouters struct{ StructureErr }

func NewErrRouters(section, reason string, err error) *ErrRouters {
	return &ErrRouters{StructureErr{section, reason, err}}
}

func (e *ErrRouters) Error() string {
	if e == nil {
		return nilStr
	}

	s := "not specified: " + e.Section
	if e.Reason != "" {
		s += " " + e.Reason
	}

	if e.Err != nil {
		s += ": " + e.Err.Error()
	}
	return s
}

// ErrNotSupported says something is still not supported
type ErrNotSupported string

func (n ErrNotSupported) Error() string {
	return fmt.Sprintf("%s is not supported", string(n))
}
