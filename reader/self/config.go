// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package self

import (
	"fmt"
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/tools"
	"github.com/pkg/errors"
)

// Config holds the necessary configuration for setting up an self reading
// facility, which is the way to record the app's metrics.
type Config struct {
	log          tools.FieldLogger
	SelfName     string
	SelfTypeName string `mapstructure:"type_name"`
	SelfInterval string `mapstructure:"interval"`
	SelfBackoff  int    `mapstructure:"backoff"`
	SelfEndpoint string // this is for testing purposes and you are not supposed to set it
	mapper       datatype.Mapper
	Cinterval    time.Duration
}

// Conf func is used for initializing a Config object.
type Conf func(*Config) error

// NewConfig returns an instance of the expvar reader.
func NewConfig(conf ...Conf) (*Config, error) {
	obj := new(Config)
	for _, c := range conf {
		err := c(obj)
		if err != nil {
			return nil, err
		}
	}
	if obj.mapper == nil {
		obj.mapper = datatype.DefaultMapper()
	}
	return obj, nil
}

// Reader implements the RecorderConf interface.
func (c *Config) Reader() (reader.DataReader, error) {
	return New(
		reader.WithLogger(c.Logger()),
		reader.WithEndpoint(c.Endpoint()),
		reader.WithMapper(c.mapper),
		reader.WithName(c.Name()),
		reader.WithTypeName(c.TypeName()),
		reader.WithInterval(c.Interval()),
		reader.WithTimeout(c.Timeout()),
		reader.WithBackoff(c.Backoff()),
	)
}

// Name returns the name.
func (c *Config) Name() string { return c.SelfName }

// TypeName returns the typeName.
func (c *Config) TypeName() string { return c.SelfTypeName }

// Endpoint returns the endpoint.
func (c *Config) Endpoint() string { return c.SelfEndpoint }

// Interval returns the interval.
func (c *Config) Interval() time.Duration { return c.Cinterval }

// Timeout returns the timeout.
func (c *Config) Timeout() time.Duration { return time.Second }

// Logger returns the logger.
func (c *Config) Logger() tools.FieldLogger { return c.log }

// Backoff returns the backoff.
func (c *Config) Backoff() int { return c.SelfBackoff }

// WithLogger produces an error if the log is nil.
func WithLogger(log tools.FieldLogger) Conf {
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
	AllKeys() []string
}

// WithViper produces an error any of the inputs are empty.
func WithViper(v unmarshaller, name, key string) Conf {
	return func(c *Config) error {
		var interval time.Duration
		if v == nil {
			return errors.New("no config file")
		}
		err := v.UnmarshalKey(key, &c)
		if err != nil || v.AllKeys() == nil {
			return errors.Wrap(err, "decoding config")
		}
		if interval, err = time.ParseDuration(c.SelfInterval); err != nil {
			return errors.Wrapf(err, "parse interval (%v)", c.SelfInterval)
		}
		c.Cinterval = interval
		if c.SelfTypeName == "" {
			return fmt.Errorf("type_name cannot be empty: %s", c.SelfTypeName)
		}
		c.SelfName = name
		c.mapper = datatype.DefaultMapper()
		c.SelfEndpoint = "http://127.0.0.1:9200"
		return nil
	}
}
