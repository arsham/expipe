// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvar

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	"github.com/spf13/viper"
)

// Config describes how expvar read is setup
// IMPORTANT NOTE: This file was copied to elasticsearch's config. Please create tests for that one if this API changes.

// Config holds the necessary configuration for setting up an expvar reader endpoint.
type Config struct {
	name       string
	Endpoint_  string `mapstructure:"endpoint"`
	RoutePath_ string `mapstructure:"routepath"`
	Interval_  string `mapstructure:"interval"`
	Timeout_   string `mapstructure:"timeout"`
	LogLevel_  string `mapstructure:"log_level"`
	Backoff_   int    `mapstructure:"backoff"`

	logger   logrus.FieldLogger
	interval time.Duration
	timeout  time.Duration
}

func NewConfig(name string, log logrus.FieldLogger, endpoint, routepath string, interval, timeout time.Duration, backoff int) (*Config, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}
	url, err := lib.SanitiseURL(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %d", endpoint)
	}

	return &Config{
		name:       name,
		Endpoint_:  url,
		RoutePath_: routepath,
		timeout:    timeout,
		interval:   interval,
		logger:     log,
		Backoff_:   backoff,
	}, nil
}

// FromViper constructs the necessary configuration for bootstrapping the expvar reader
func FromViper(v *viper.Viper, name, key string) (*Config, error) {
	var (
		c         Config
		inter, to time.Duration
	)
	err := v.UnmarshalKey(key, &c)
	if err != nil {
		return nil, fmt.Errorf("decodeing config: %s", err)
	}
	if inter, err = time.ParseDuration(c.Interval_); err != nil {
		return nil, fmt.Errorf("parse interval (%v): %s", c.Interval_, err)
	}
	if to, err = time.ParseDuration(c.Timeout_); err != nil {
		return nil, fmt.Errorf("parse timeout: %s", err)
	}
	if c.Backoff_ <= 5 {
		return nil, fmt.Errorf("back off should be at least 5: %d", c.Backoff_)
	}
	c.interval, c.timeout = inter, to

	c.logger = logrus.StandardLogger()
	if c.LogLevel_ != "" {
		c.logger = lib.GetLogger(c.LogLevel_)
	}

	if c.Endpoint_ == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}
	url, err := lib.SanitiseURL(c.Endpoint_)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %d", c.Endpoint_)
	}
	c.Endpoint_ = url

	c.name = name
	return &c, nil
}

func (c *Config) NewInstance(ctx context.Context) (reader.DataReader, error) {
	endpoint, err := url.Parse(c.Endpoint())
	if err != nil {
		return nil, err
	}
	endpoint.Path = path.Join(endpoint.Path, c.RoutePath())
	ctxReader := reader.NewCtxReader(endpoint.String())
	return NewExpvarReader(c.logger, ctxReader, c.name, c.interval, c.timeout)
}

func (c *Config) Name() string               { return c.name }
func (c *Config) Endpoint() string           { return c.Endpoint_ }
func (c *Config) RoutePath() string          { return c.RoutePath_ }
func (c *Config) Interval() time.Duration    { return c.interval }
func (c *Config) Timeout() time.Duration     { return c.timeout }
func (c *Config) Logger() logrus.FieldLogger { return c.logger }
func (c *Config) Backoff() int               { return c.Backoff_ }
