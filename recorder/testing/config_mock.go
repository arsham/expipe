// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"time"

	"github.com/arsham/expvastic/internal"
	"github.com/arsham/expvastic/recorder"
)

// Config holds the necessary configuration for setting up an elasticsearch reader endpoint.
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
func (m *Config) NewInstance(ctx context.Context) (recorder.DataRecorder, error) {
	return New(ctx, m.Logger(), m.Name(), m.Endpoint(), m.IndexName(), m.Timeout(), m.Backoff())
}

// Name is the mocked version
func (m *Config) Name() string { return m.MockName }

// IndexName is the mocked version
func (m *Config) IndexName() string { return m.MockIndexName }

// Endpoint is the mocked version
func (m *Config) Endpoint() string { return m.MockEndpoint }

// RoutePath is the mocked version
func (m *Config) RoutePath() string { return "" }

// Timeout is the mocked version
func (m *Config) Timeout() time.Duration { return m.MockTimeout }

// Logger is the mocked version
func (m *Config) Logger() internal.FieldLogger { return m.MockLogger }

// Backoff is the mocked version
func (m *Config) Backoff() int { return m.MockBackoff }
