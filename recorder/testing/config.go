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

// NewConfig returns a mocked object
func NewConfig(name string, log internal.FieldLogger, endpoint string, timeout time.Duration, backoff int, indexName string) (*Config, error) {
	return &Config{
		MockName:      name,
		MockEndpoint:  endpoint,
		MockTimeout:   timeout,
		MockBackoff:   backoff,
		MockIndexName: indexName,
		MockLogger:    log,
	}, nil
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
