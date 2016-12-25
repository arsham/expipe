// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
	"context"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/communication"
)

// MockConfig holds the necessary configuration for setting up an elasticsearch reader endpoint.
type MockConfig struct {
	MockName      string
	MockEndpoint  string
	MockTimeout   time.Duration
	MockBackoff   int
	MockIndexName string
	MockLogger    logrus.FieldLogger
}

// NewMockConfig returns a mocked object
func NewMockConfig(name string, log logrus.FieldLogger, endpoint string, timeout time.Duration, backoff int, indexName string) (*MockConfig, error) {
	return &MockConfig{
		MockName:      name,
		MockEndpoint:  endpoint,
		MockTimeout:   timeout,
		MockBackoff:   backoff,
		MockIndexName: indexName,
		MockLogger:    log,
	}, nil
}

// NewInstance returns a mocked object
func (m *MockConfig) NewInstance(ctx context.Context, payloadChan chan *RecordJob, errorChan chan<- communication.ErrorMessage) (DataRecorder, error) {
	return NewSimpleRecorder(ctx, m.Logger(), payloadChan, errorChan, m.Name(), m.Endpoint(), m.IndexName(), m.Timeout())
}

// Name is the mocked version
func (m *MockConfig) Name() string { return m.MockName }

// IndexName is the mocked version
func (m *MockConfig) IndexName() string { return m.MockIndexName }

// Endpoint is the mocked version
func (m *MockConfig) Endpoint() string { return m.MockEndpoint }

// RoutePath is the mocked version
func (m *MockConfig) RoutePath() string { return "" }

// Timeout is the mocked version
func (m *MockConfig) Timeout() time.Duration { return m.MockTimeout }

// Logger is the mocked version
func (m *MockConfig) Logger() logrus.FieldLogger { return m.MockLogger }

// Backoff is the mocked version
func (m *MockConfig) Backoff() int { return m.MockBackoff }
