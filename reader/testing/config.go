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

// NewConfig returns a mocked version of the Config
func NewConfig(name, typeName string, log internal.FieldLogger, endpoint string, interval, timeout time.Duration, backoff int) (*Config, error) {
	return &Config{
		MockName:     name,
		MockTypeName: typeName,
		MockEndpoint: endpoint,
		MockTimeout:  timeout,
		MockInterval: interval,
		MockLogger:   log,
		MockBackoff:  backoff,
	}, nil
}

// NewInstance  returns a mocked version of the config
func (c *Config) NewInstance() (reader.DataReader, error) {
	return New(
		reader.SetLogger(c.Logger()),
		reader.SetEndpoint(c.Endpoint()),
		reader.SetName(c.Name()),
		reader.SetTypeName(c.TypeName()),
		reader.SetInterval(c.Interval()),
		reader.SetTimeout(c.Timeout()),
		reader.SetBackoff(c.Backoff()),
	)
}

// Name returns the name
func (c *Config) Name() string { return c.MockName }

// TypeName returns the typename
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
