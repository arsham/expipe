// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package elasticsearch

import (
	"context"
	"time"

	"github.com/arsham/expvastic/internal"
	"github.com/arsham/expvastic/recorder"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Config holds the necessary configuration for setting up an elasticsearch reader endpoint.
type Config struct {
	name        string
	ESEndpoint  string `mapstructure:"endpoint"`
	ESTimeout   string `mapstructure:"timeout"`
	ESBackoff   int    `mapstructure:"backoff"`
	ESIndexName string `mapstructure:"index_name"`

	log     internal.FieldLogger
	timeout time.Duration
}

// NewConfig returns errors coming from Viper
func NewConfig(log internal.FieldLogger, name string, endpoint string, timeout time.Duration, backoff int, indexName string) (*Config, error) {
	c := &Config{
		name:        name,
		ESEndpoint:  endpoint,
		timeout:     timeout,
		ESBackoff:   backoff,
		ESIndexName: indexName,
		log:         log,
	}
	return withConfig(c)
}

// FromViper constructs the necessary configuration for bootstrapping the elasticsearch reader
func FromViper(v *viper.Viper, log internal.FieldLogger, name, key string) (*Config, error) {
	var (
		c       Config
		timeout time.Duration
	)
	err := v.UnmarshalKey(key, &c)
	if err != nil {
		return nil, errors.Wrap(err, "decoding config")
	}

	c.name = name
	c.log = log

	if timeout, err = time.ParseDuration(c.ESTimeout); err != nil {
		return nil, &recorder.ErrParseTimeOut{Timeout: c.ESTimeout, Err: err}
	}
	c.timeout = timeout

	return withConfig(&c)
}

func withConfig(c *Config) (*Config, error) {
	if c.name == "" {
		return nil, recorder.ErrEmptyName
	}

	if c.ESEndpoint == "" {
		return nil, recorder.ErrEmptyEndpoint
	}
	endpoint, err := internal.SanitiseURL(c.ESEndpoint)
	if err != nil {
		return nil, recorder.ErrInvalidEndpoint(c.ESEndpoint)
	}
	c.ESEndpoint = endpoint

	if c.ESIndexName == "" {
		return nil, recorder.ErrEmptyIndexName
	}

	return c, nil
}

// NewInstance returns an instance of the elasticsearch recorder
func (c *Config) NewInstance(ctx context.Context) (recorder.DataRecorder, error) {
	return New(ctx, c.Logger(), c.Name(), c.Endpoint(), c.IndexName(), c.Timeout(), c.Backoff())
}

// Name return the name
func (c *Config) Name() string { return c.name }

// IndexName return the index name
func (c *Config) IndexName() string { return c.ESIndexName }

// Endpoint return the endpoint
func (c *Config) Endpoint() string { return c.ESEndpoint }

// RoutePath return an empty string
func (c *Config) RoutePath() string { return "" }

// Timeout return the timeout
func (c *Config) Timeout() time.Duration { return c.timeout }

// Logger return the logger
func (c *Config) Logger() internal.FieldLogger { return c.log }

// Backoff return the backoff
func (c *Config) Backoff() int { return c.ESBackoff }
