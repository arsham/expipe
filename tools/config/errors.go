// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config

import "fmt"

// ErrEmptyConfig is returned when the config file is empty.
var ErrEmptyConfig = &StructureErr{"", "empty configuration file", nil}

const (
	// NilStr is used to print nil string for an error.
	NilStr = "<nil>"
)

// StructureErr is for reading the configuration file errors.
type StructureErr struct {
	Section string // The section that error happened
	Reason  string // The reason behind the error
	Err     error  // Err is the occurred error during the operation.
}

// Error returns "<nil>" if the error is nil.
func (e *StructureErr) Error() string {
	if e == nil {
		return NilStr
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

// NotSpecifiedError is returned when a section is not specified.
type NotSpecifiedError StructureErr

// NewNotSpecifiedError returns ErrNotSpecified with the given input.
func NewNotSpecifiedError(section, reason string, err error) *NotSpecifiedError {
	return &NotSpecifiedError{section, reason, err}
}

func (e *NotSpecifiedError) Error() string {
	if e == nil {
		return NilStr
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

// RoutersError is returned when routes are not configured correctly. The
// section on this error is the subsection of the route.
type RoutersError struct{ StructureErr }

// NewRoutersError instantiates an RoutersError with the given input.
func NewRoutersError(section, reason string, err error) *RoutersError {
	return &RoutersError{StructureErr{section, reason, err}}
}

func (e *RoutersError) Error() string {
	if e == nil {
		return NilStr
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

// NotSupportedError says something is still not supported.
type NotSupportedError string

func (n NotSupportedError) Error() string {
	return fmt.Sprintf("%s is not supported", string(n))
}
