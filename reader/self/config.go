// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package self

import (
	"context"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/reader"
	"github.com/spf13/viper"
)

// Config holds the necessary configuration for setting up an self reading facility.
type Config struct {
	name         string
	SelfTypeName string `mapstructure:"type_name"`
	SelfInterval string `mapstructure:"interval"`
	SelfBackoff  int    `mapstructure:"backoff"`
	SelfEndpoint string // this is for testing purposes and you are not supposed to set it
	mapper       datatype.Mapper

	interval time.Duration
	log      logrus.FieldLogger
}

// FromViper constructs the necessary configuration for bootstrapping the expvar reader
func FromViper(v *viper.Viper, log logrus.FieldLogger, name, key string) (*Config, error) {
	var (
		c     Config
		inter time.Duration
	)
	err := v.UnmarshalKey(key, &c)
	if err != nil {
		return nil, fmt.Errorf("decoding config: %s", err)
	}
	if inter, err = time.ParseDuration(c.SelfInterval); err != nil {
		return nil, fmt.Errorf("parse interval (%v): %s", c.SelfInterval, err)
	}

	if c.SelfTypeName == "" {
		return nil, fmt.Errorf("type_name cannot be empty: %s", c.SelfTypeName)
	}

	c.mapper = datatype.DefaultMapper()
	c.interval = inter
	c.log = log
	c.name = name
	c.SelfEndpoint = "http://127.0.0.1:9200"
	return &c, nil
}

// NewInstance instantiates a SelfReader
func (c *Config) NewInstance(ctx context.Context) (reader.DataReader, error) {
	return New(c.Logger(), c.Endpoint(), c.mapper, c.Name(), c.TypeName(), c.Interval(), c.Timeout(), c.Backoff())
}

// Name returns the name
func (c *Config) Name() string { return c.name }

// TypeName returns the typename
func (c *Config) TypeName() string { return c.SelfTypeName }

// Endpoint returns the endpoint
func (c *Config) Endpoint() string { return c.SelfEndpoint }

// RoutePath returns the routepath
func (c *Config) RoutePath() string { return "" }

// Interval returns the interval
func (c *Config) Interval() time.Duration { return c.interval }

// Timeout returns the timeout
func (c *Config) Timeout() time.Duration { return time.Second }

// Logger returns the logger
func (c *Config) Logger() logrus.FieldLogger { return c.log }

// Backoff returns the backoff
func (c *Config) Backoff() int { return c.SelfBackoff }
