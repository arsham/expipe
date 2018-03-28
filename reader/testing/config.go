// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/reader"
)

// Config is used for instantiating a mock reader
type Config struct {
	MockName     string
	MockTypeName string
	MockEndpoint string
	MockTimeout  time.Duration
	MockInterval time.Duration
	MockBackoff  int
	MockLogger   internal.FieldLogger
}

// Conf func is used for initializing a Config object.
type Conf func(*Config) error

// NewConfig returns a mocked object
func NewConfig(conf ...Conf) (*Config, error) {
	obj := new(Config)
	for _, c := range conf {
		c(obj)
	}
	return obj, nil
}

// NewInstance  returns a mocked version of the config
func (c *Config) NewInstance() (reader.DataReader, error) {
	return New(
		reader.WithLogger(c.Logger()),
		reader.WithEndpoint(c.Endpoint()),
		reader.WithName(c.Name()),
		reader.WithTypeName(c.TypeName()),
		reader.WithInterval(c.Interval()),
		reader.WithTimeout(c.Timeout()),
		reader.WithBackoff(c.Backoff()),
	)
}

// Name returns the name
func (c *Config) Name() string { return c.MockName }

// TypeName returns the typeName
func (c *Config) TypeName() string { return c.MockTypeName }

// Endpoint returns the endpoint
func (c *Config) Endpoint() string { return c.MockEndpoint }

// Interval returns the interval
func (c *Config) Interval() time.Duration { return c.MockInterval }

// Timeout returns the timeout
func (c *Config) Timeout() time.Duration { return c.MockTimeout }

// Logger returns the logger
func (c *Config) Logger() internal.FieldLogger { return c.MockLogger }

// Backoff returns the backoff
func (c *Config) Backoff() int { return c.MockBackoff }

// WithLogger doesn't produce any errors.
func WithLogger(log internal.FieldLogger) Conf {
	return func(c *Config) error {
		c.MockLogger = log
		return nil
	}
}

// WithName doesn't produce any errors.
func WithName(name string) Conf {
	return func(c *Config) error {
		c.MockName = name
		return nil
	}
}

// WithTypeName doesn't produce any errors.
func WithTypeName(typeName string) Conf {
	return func(c *Config) error {
		c.MockTypeName = typeName
		return nil
	}
}

// WithEndpoint doesn't produce any errors.
func WithEndpoint(endpoint string) Conf {
	return func(c *Config) error {
		c.MockEndpoint = endpoint
		return nil
	}
}

// WithTimeout doesn't produce any errors.
func WithTimeout(timeout time.Duration) Conf {
	return func(c *Config) error {
		c.MockTimeout = timeout
		return nil
	}
}

// WithInterval doesn't produce any errors.
func WithInterval(internal time.Duration) Conf {
	return func(c *Config) error {
		c.MockInterval = internal
		return nil
	}
}

// WithBackoff doesn't produce any errors.
func WithBackoff(backoff int) Conf {
	return func(c *Config) error {
		c.MockBackoff = backoff
		return nil
	}
}
