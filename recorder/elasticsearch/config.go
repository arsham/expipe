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

// Conf func is used for initializing a Config object.
type Conf func(*Config) error

// NewConfig returns any errors that any of conf function return.
func NewConfig(conf ...Conf) (*Config, error) {
	obj := new(Config)
	for _, c := range conf {
		err := c(obj)
		if err != nil {
			return nil, err
		}
	}

	if obj.name == "" {
		return nil, recorder.ErrEmptyName
	}
	if obj.ESEndpoint == "" {
		return nil, recorder.ErrEmptyEndpoint
	}
	if obj.ESIndexName == "" {
		return nil, recorder.ErrEmptyIndexName
	}
	if obj.ESBackoff == 0 {
		obj.ESBackoff = 5
	}
	if obj.timeout == 0 {
		obj.timeout = time.Second
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
func (c *Config) Name() string { return c.name }

// IndexName return the index name
func (c *Config) IndexName() string { return c.ESIndexName }

// Endpoint return the endpoint
func (c *Config) Endpoint() string { return c.ESEndpoint }

// Timeout return the timeout
func (c *Config) Timeout() time.Duration { return c.timeout }

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

// WithName produces an error if name is empty.
func WithName(name string) Conf {
	return func(c *Config) error {
		if name == "" {
			return recorder.ErrEmptyName
		}
		c.name = name
		return nil
	}
}

// WithIndexName produces an error if indexName is empty.
func WithIndexName(indexName string) Conf {
	return func(c *Config) error {
		if indexName == "" {
			return recorder.ErrInvalidIndexName(indexName)
		}
		c.ESIndexName = indexName
		return nil
	}
}

// WithEndpoint produces an error if endpoint is empty.
func WithEndpoint(endpoint string) Conf {
	return func(c *Config) error {
		if endpoint == "" {
			return recorder.ErrEmptyEndpoint
		}
		endpoint, err := internal.SanitiseURL(endpoint)
		if err != nil {
			return recorder.ErrInvalidEndpoint(endpoint)
		}
		c.ESEndpoint = endpoint
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
			return &recorder.ErrParseTimeOut{Timeout: c.ESTimeout, Err: err}
		}
		c.name = name
		c.timeout = timeout
		return nil
	}
}

// WithTimeout doesn't produce any errors.
func WithTimeout(timeout time.Duration) Conf {
	return func(c *Config) error {
		c.timeout = timeout
		return nil
	}
}

// WithBackoff doesn't produce any errors.
func WithBackoff(backoff int) Conf {
	return func(c *Config) error {
		c.ESBackoff = backoff
		return nil
	}
}
