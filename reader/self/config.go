// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package self

import (
	"context"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/reader"
	"github.com/spf13/viper"
)

// Config describes how expvastic's own app is setup
// IMPORTANT NOTE: This file was copied to elasticsearch's config. Please create tests for that one if this API changes.

// Config holds the necessary configuration for setting up an self reading facility.
type Config struct {
	name      string
	TypeName_ string `mapstructure:"type_name"`
	Interval_ string `mapstructure:"interval"`

	interval time.Duration
	logger   logrus.FieldLogger
}

// FromViper constructs the necessary configuration for bootstrapping the expvar reader
func FromViper(v *viper.Viper, name, key string) (*Config, error) {
	var (
		c     Config
		inter time.Duration
	)
	err := v.UnmarshalKey(key, &c)
	if err != nil {
		return nil, fmt.Errorf("decodeing config: %s", err)
	}
	if inter, err = time.ParseDuration(c.Interval_); err != nil {
		return nil, fmt.Errorf("parse interval (%v): %s", c.Interval_, err)
	}

	if c.TypeName_ == "" {
		return nil, fmt.Errorf("type_name cannot be empty: %s", c.TypeName_)
	}

	c.interval = inter
	c.logger = logrus.StandardLogger()
	c.name = name
	return &c, nil
}

func (c *Config) NewInstance(ctx context.Context) (reader.DataReader, error) {
	return NewSelfReader(c.logger, c.name, c.TypeName(), c.interval)
}

func (c *Config) Name() string               { return c.name }
func (c *Config) TypeName() string           { return c.TypeName_ }
func (c *Config) Endpoint() string           { return "" }
func (c *Config) RoutePath() string          { return "" }
func (c *Config) Interval() time.Duration    { return c.interval }
func (c *Config) Timeout() time.Duration     { return 0 }
func (c *Config) Logger() logrus.FieldLogger { return c.logger }
func (c *Config) Backoff() int               { return 0 }
