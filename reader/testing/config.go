// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"time"

	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/tools"
)

// Config is used for instantiating a mock reader.
type Config struct {
	MockName     string
	MockTypeName string
	MockEndpoint string
	MockTimeout  time.Duration
	MockInterval time.Duration
	MockLogger   tools.FieldLogger
}

// Reader implements the ReaderConf interface.
func (c *Config) Reader() (reader.DataReader, error) {
	return New(
		reader.WithLogger(c.Logger()),
		reader.WithEndpoint(c.Endpoint()),
		reader.WithName(c.Name()),
		reader.WithTypeName(c.TypeName()),
		reader.WithInterval(c.Interval()),
		reader.WithTimeout(c.Timeout()),
	)
}

// Name returns the name.
func (c *Config) Name() string { return c.MockName }

// TypeName returns the typeName.
func (c *Config) TypeName() string { return c.MockTypeName }

// Endpoint returns the endpoint.
func (c *Config) Endpoint() string { return c.MockEndpoint }

// Interval returns the interval.
func (c *Config) Interval() time.Duration { return c.MockInterval }

// Timeout returns the timeout.
func (c *Config) Timeout() time.Duration { return c.MockTimeout }

// Logger returns the logger.
func (c *Config) Logger() tools.FieldLogger { return c.MockLogger }
