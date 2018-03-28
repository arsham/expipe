// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/recorder"
)

// Config holds the necessary configuration for setting up an elasticsearch recorder endpoint.
type Config struct {
	MockName      string
	MockEndpoint  string
	MockTimeout   time.Duration
	MockBackoff   int
	MockIndexName string
	MockLogger    internal.FieldLogger
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

// NewInstance returns a mocked object
func (c *Config) NewInstance() (recorder.DataRecorder, error) {
	return New(
		recorder.WithLogger(c.Logger()),
		recorder.WithEndpoint(c.Endpoint()),
		recorder.WithName(c.Name()),
		recorder.WithIndexName(c.IndexName()),
		recorder.WithTimeout(c.Timeout()),
		recorder.WithBackoff(c.Backoff()),
	)
}

// Name is the mocked version
func (c *Config) Name() string { return c.MockName }

// IndexName is the mocked version
func (c *Config) IndexName() string { return c.MockIndexName }

// Endpoint is the mocked version
func (c *Config) Endpoint() string { return c.MockEndpoint }

// Timeout is the mocked version
func (c *Config) Timeout() time.Duration { return c.MockTimeout }

// Logger is the mocked version
func (c *Config) Logger() internal.FieldLogger { return c.MockLogger }

// Backoff is the mocked version
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

// WithIndexName doesn't produce any errors.
func WithIndexName(indexName string) Conf {
	return func(c *Config) error {
		c.MockIndexName = indexName
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

// WithBackoff doesn't produce any errors.
func WithBackoff(backoff int) Conf {
	return func(c *Config) error {
		c.MockBackoff = backoff
		return nil
	}
}
