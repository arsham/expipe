// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
    "context"
    "time"

    "github.com/Sirupsen/logrus"
)

// MockConfig holds the necessary configuration for setting up an elasticsearch reader endpoint.
type MockConfig struct {
    Name_      string
    Endpoint_  string
    Timeout_   time.Duration
    Interval_  time.Duration
    Backoff_   int
    IndexName_ string
    Logger_    logrus.FieldLogger
}

func NewMockConfig(name string, log logrus.FieldLogger, endpoint string, interval, timeout time.Duration, backoff int, indexName string) (*MockConfig, error) {
    return &MockConfig{
        Name_:      name,
        Endpoint_:  endpoint,
        Timeout_:   timeout,
        Interval_:  interval,
        Backoff_:   backoff,
        IndexName_: indexName,
        Logger_:    log,
    }, nil
}

func (m *MockConfig) NewInstance(ctx context.Context, payloadChan chan *RecordJob) (DataRecorder, error) {
    return NewSimpleRecorder(ctx, m.Logger(), payloadChan, m.Name(), m.Endpoint(), m.IndexName(), m.Interval(), m.Timeout())
}
func (m *MockConfig) Name() string               { return m.Name_ }
func (m *MockConfig) IndexName() string          { return m.IndexName_ }
func (m *MockConfig) Endpoint() string           { return m.Endpoint_ }
func (m *MockConfig) RoutePath() string          { return "" }
func (m *MockConfig) Interval() time.Duration    { return m.Interval_ }
func (m *MockConfig) Timeout() time.Duration     { return m.Timeout_ }
func (m *MockConfig) Logger() logrus.FieldLogger { return m.Logger_ }
func (m *MockConfig) Backoff() int               { return m.Backoff_ }
