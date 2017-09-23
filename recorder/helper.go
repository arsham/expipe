// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
	"strings"
	"time"

	"github.com/arsham/expipe/internal"
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

	// SetIndexName is for setting the IndexName
	SetIndexName(indexName string)

	// SetEndpoint is for setting the Endpoint
	SetEndpoint(endpoint string)

	// SetTimeout is for setting the Timeout
	SetTimeout(timeout time.Duration)

	// SetBackoff is for setting the Backoff
	SetBackoff(backoff int)
}

// SetLogger sets the log of the recorder
func SetLogger(log internal.FieldLogger) func(Constructor) error {
	return func(e Constructor) error {
		if log == nil {
			return errors.New("recorder nil logger")
		}
		e.SetLogger(log)
		return nil
	}
}

// SetName sets the name of the recorder
func SetName(name string) func(Constructor) error {
	return func(e Constructor) error {
		if name == "" {
			return ErrEmptyName
		}
		e.SetName(name)
		return nil
	}
}

// SetEndpoint sets the endpoint of the recorder
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

// SetIndexName sets the indexName of the recorder
func SetIndexName(indexName string) func(Constructor) error {
	return func(e Constructor) error {
		if indexName == "" {
			return ErrEmptyIndexName
		}
		if strings.ContainsAny(indexName, ` "*\<|,>/?`) {
			return ErrInvalidIndexName(indexName)
		}
		e.SetIndexName(indexName)
		return nil
	}
}

// SetTimeout sets the timeout of the recorder
func SetTimeout(timeout time.Duration) func(Constructor) error {
	return func(e Constructor) error {
		if timeout < time.Second {
			return ErrLowTimeout(timeout)
		}
		e.SetTimeout(timeout)
		return nil
	}
}

// SetBackoff sets the backoff of the recorder
func SetBackoff(backoff int) func(Constructor) error {
	return func(e Constructor) error {
		if backoff < 5 {
			return ErrLowBackoffValue(backoff)
		}

		e.SetBackoff(backoff)
		return nil
	}
}
