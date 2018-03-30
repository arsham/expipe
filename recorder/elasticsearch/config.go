// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package elasticsearch

import (
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/recorder"
	"github.com/pkg/errors"
)

// Config holds the necessary configuration for setting up an elasticsearch
// reader endpoint from a configuration file.
type Config struct {
	ESEndpoint  string `mapstructure:"endpoint"`
	ESTimeout   string `mapstructure:"timeout"`
	ESBackoff   int    `mapstructure:"backoff"`
	ESIndexName string `mapstructure:"index_name"`
	log         internal.FieldLogger
	ESName      string
	ConfTimeout time.Duration
}

// Conf func is used for initializing a Config object.
type Conf func(*Config) error

// NewConfig is used for returning the values from config file.
// It returns any errors that any of conf function return.
func NewConfig(conf ...Conf) (*Config, error) {
	obj := new(Config)
	for _, c := range conf {
		err := c(obj)
		if err != nil {
			return nil, err
		}
	}
	return obj, nil
}

// NewInstance returns an instance of the elasticsearch recorder
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

// Name return the name
func (c *Config) Name() string { return c.ESName }

// IndexName return the index name
func (c *Config) IndexName() string { return c.ESIndexName }

// Endpoint return the endpoint
func (c *Config) Endpoint() string { return c.ESEndpoint }

// Timeout return the timeout
func (c *Config) Timeout() time.Duration { return c.ConfTimeout }

// Logger return the logger
func (c *Config) Logger() internal.FieldLogger { return c.log }

// Backoff return the backoff
func (c *Config) Backoff() int { return c.ESBackoff }

// WithLogger produces an error if the log is nil.
func WithLogger(log internal.FieldLogger) Conf {
	return func(c *Config) error {
		if log == nil {
			return errors.New("nil logger")
		}
		c.log = log
		return nil
	}
}

type unmarshaller interface {
	UnmarshalKey(key string, rawVal interface{}) error
}

// WithViper produces an error any of the inputs are empty.
func WithViper(v unmarshaller, name, key string) Conf {
	return func(c *Config) error {
		if name == "" {
			return recorder.ErrEmptyName
		}
		if key == "" {
			return errors.New("key cannot be empty")
		}
		if v == nil {
			return errors.New("no config file")
		}

		var timeout time.Duration
		err := v.UnmarshalKey(key, &c)
		if err != nil {
			return errors.Wrap(err, "decoding config")
		}
		if timeout, err = time.ParseDuration(c.ESTimeout); err != nil {
			return &recorder.ParseTimeOutError{Timeout: c.ESTimeout, Err: err}
		}
		c.ESName = name
		c.ConfTimeout = timeout
		return nil
	}
}
