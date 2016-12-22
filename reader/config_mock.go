// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"context"
	"time"

	"github.com/Sirupsen/logrus"
)

type MockConfig struct {
	Name_      string
	TypeName_  string
	Endpoint_  string
	RoutePath_ string
	Timeout_   time.Duration
	Interval_  time.Duration
	Backoff_   int
	Logger_    logrus.FieldLogger
}

func NewMockConfig(name, typeName string, log logrus.FieldLogger, endpoint, routepath string, interval, timeout time.Duration, backoff int) (*MockConfig, error) {
	return &MockConfig{
		Name_:      name,
		TypeName_:  typeName,
		Endpoint_:  endpoint,
		RoutePath_: routepath,
		Timeout_:   timeout,
		Interval_:  interval,
		Logger_:    log,
		Backoff_:   backoff,
	}, nil
}

func (c *MockConfig) NewInstance(ctx context.Context, jobChan chan context.Context, resultChan chan *ReadJobResult) (DataReader, error) {
	ctxReader := NewCtxReader(c.Endpoint())
	return NewSimpleReader(c.Logger_, ctxReader, jobChan, resultChan, c.Name_, c.TypeName_, c.Interval_, c.Timeout_)
}

func (c *MockConfig) Name() string               { return c.Name_ }
func (c *MockConfig) TypeName() string           { return c.TypeName_ }
func (c *MockConfig) Endpoint() string           { return c.Endpoint_ }
func (c *MockConfig) RoutePath() string          { return c.RoutePath_ }
func (c *MockConfig) Interval() time.Duration    { return c.Interval_ }
func (c *MockConfig) Timeout() time.Duration     { return c.Timeout_ }
func (c *MockConfig) Logger() logrus.FieldLogger { return c.Logger_ }
func (c *MockConfig) Backoff() int               { return c.Backoff_ }
