// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/tools"
	"github.com/pkg/errors"
)

// This file contains the construction functions required for instantiating a
// Reader object. Input variables are sanitised here.

// Constructor is a Reader object that will accept configurations.
type Constructor interface {
	SetLogger(logger tools.FieldLogger)
	SetName(name string)
	SetTypeName(typeName string)
	SetEndpoint(endpoint string)
	SetMapper(mapper datatype.Mapper)
	SetInterval(interval time.Duration)
	SetTimeout(timeout time.Duration)
}

// WithLogger sets the log of the reader.
func WithLogger(log tools.FieldLogger) func(Constructor) error {
	return func(e Constructor) error {
		if log == nil {
			return ErrNillLogger
		}
		e.SetLogger(log)
		return nil
	}
}

// WithName sets the name of the reader.
func WithName(name string) func(Constructor) error {
	return func(e Constructor) error {
		if name == "" {
			return ErrEmptyName
		}
		e.SetName(name)
		return nil
	}
}

// WithEndpoint sets the endpoint of the reader.
func WithEndpoint(endpoint string) func(Constructor) error {
	return func(e Constructor) error {
		if endpoint == "" {
			return ErrEmptyEndpoint
		}
		url, err := tools.SanitiseURL(endpoint)
		if err != nil {
			return InvalidEndpointError(endpoint)
		}
		e.SetEndpoint(url)
		return nil
	}
}

// WithMapper sets the mapper of the reader.
func WithMapper(mapper datatype.Mapper) func(Constructor) error {
	return func(e Constructor) error {
		if mapper == nil {
			return errors.New("nil mapper")
		}
		e.SetMapper(mapper)
		return nil
	}
}

// WithTypeName sets the typeName of the reader.
func WithTypeName(typeName string) func(Constructor) error {
	return func(e Constructor) error {
		if typeName == "" {
			return ErrEmptyTypeName
		}
		e.SetTypeName(typeName)
		return nil
	}
}

// WithInterval sets the interval of the reader.
func WithInterval(interval time.Duration) func(Constructor) error {
	return func(e Constructor) error {
		if interval == time.Duration(0) {
			return LowIntervalError(interval)
		}
		e.SetInterval(interval)
		return nil
	}
}

// WithTimeout sets the timeout of the reader.
func WithTimeout(timeout time.Duration) func(Constructor) error {
	return func(e Constructor) error {
		if timeout < time.Second {
			return LowTimeoutError(timeout)
		}
		e.SetTimeout(timeout)
		return nil
	}
}
