// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/internal/datatype"
	"github.com/pkg/errors"
)

// This file contains the construction functions required for instantiating
// a Reader object. Input variables are sanitised here.

// Constructor is an interface for setting up an object for testing.
type Constructor interface {
	// SetLogger is for setting the Logger
	SetLogger(logger internal.FieldLogger)

	// SetName is for setting the Name
	SetName(name string)

	// SetTypeName is for setting the TypeName
	SetTypeName(typeName string)

	// SetEndpoint is for setting the Endpoint
	SetEndpoint(endpoint string)

	// SetMapper is for setting the SetMapper
	SetMapper(mapper datatype.Mapper)

	// SetInterval is for setting the Interval
	SetInterval(interval time.Duration)

	// SetTimeout is for setting the Timeout
	SetTimeout(timeout time.Duration)

	// SetBackoff is for setting the Backoff
	SetBackoff(backoff int)
}

// SetLogger sets the log of the reader
func SetLogger(log internal.FieldLogger) func(Constructor) error {
	return func(e Constructor) error {
		if log == nil {
			return errors.New("reader nil logger")
		}
		e.SetLogger(log)
		return nil
	}
}

// SetName sets the name of the reader
func SetName(name string) func(Constructor) error {
	return func(e Constructor) error {
		if name == "" {
			return ErrEmptyName
		}
		e.SetName(name)
		return nil
	}
}

// SetEndpoint sets the endpoint of the reader
func SetEndpoint(endpoint string) func(Constructor) error {
	return func(e Constructor) error {
		if endpoint == "" {
			return ErrEmptyEndpoint
		}
		url, err := internal.SanitiseURL(endpoint)
		if err != nil {
			return ErrInvalidEndpoint(endpoint)
		}

		e.SetEndpoint(url)
		return nil
	}
}

// SetMapper sets the mapper of the reader
func SetMapper(mapper datatype.Mapper) func(Constructor) error {
	return func(e Constructor) error {
		if mapper == nil {
			return errors.New("nil mapper")
		}
		e.SetMapper(mapper)
		return nil
	}
}

// SetTypeName sets the typeName of the reader
func SetTypeName(typeName string) func(Constructor) error {
	return func(e Constructor) error {
		if typeName == "" {
			return ErrEmptyTypeName
		}
		e.SetTypeName(typeName)
		return nil
	}
}

// SetInterval sets the interval of the reader
func SetInterval(interval time.Duration) func(Constructor) error {
	return func(e Constructor) error {
		if interval == time.Duration(0) {
			return ErrLowInterval(interval)
		}
		e.SetInterval(interval)
		return nil
	}
}

// SetTimeout sets the timeout of the reader
func SetTimeout(timeout time.Duration) func(Constructor) error {
	return func(e Constructor) error {
		if timeout < time.Second {
			return ErrLowTimeout(timeout)
		}
		e.SetTimeout(timeout)
		return nil
	}
}

// SetBackoff sets the backoff of the reader
func SetBackoff(backoff int) func(Constructor) error {
	return func(e Constructor) error {
		if backoff < 5 {
			return ErrLowBackoffValue(backoff)
		}

		e.SetBackoff(backoff)
		return nil
	}
}
