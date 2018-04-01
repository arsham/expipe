// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
	"strings"
	"time"

	"github.com/arsham/expipe/tools"
	"github.com/pkg/errors"
)

// This file contains the construction functions required for instantiating
// a Reader object. Input variables are sanitised here.

// Constructor is an interface for setting up an object for testing.
type Constructor interface {
	SetLogger(logger tools.FieldLogger)
	SetName(name string)
	SetIndexName(indexName string)
	SetEndpoint(endpoint string)
	SetTimeout(timeout time.Duration)
	SetBackoff(backoff int)
}

// WithLogger sets the log of the recorder
func WithLogger(log tools.FieldLogger) func(Constructor) error {
	return func(e Constructor) error {
		if log == nil {
			return errors.New("recorder nil logger")
		}
		e.SetLogger(log)
		return nil
	}
}

// WithName sets the name of the recorder
func WithName(name string) func(Constructor) error {
	return func(e Constructor) error {
		if name == "" {
			return ErrEmptyName
		}
		e.SetName(name)
		return nil
	}
}

// WithEndpoint sets the endpoint of the recorder
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

// WithIndexName sets the indexName of the recorder
func WithIndexName(indexName string) func(Constructor) error {
	return func(e Constructor) error {
		if indexName == "" {
			return ErrEmptyIndexName
		}
		if strings.ContainsAny(indexName, ` "*\<|,>/?`) {
			return InvalidIndexNameError(indexName)
		}
		e.SetIndexName(indexName)
		return nil
	}
}

// WithTimeout sets the timeout of the recorder
func WithTimeout(timeout time.Duration) func(Constructor) error {
	return func(e Constructor) error {
		if timeout < time.Second {
			return LowTimeout(timeout)
		}
		e.SetTimeout(timeout)
		return nil
	}
}

// WithBackoff sets the backoff of the recorder
func WithBackoff(backoff int) func(Constructor) error {
	return func(e Constructor) error {
		if backoff < 5 {
			return LowBackoffValueError(backoff)
		}
		e.SetBackoff(backoff)
		return nil
	}
}
