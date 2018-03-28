// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvar

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/reader"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Config holds the necessary configuration for setting up an expvar reader endpoint.
// If MapFile is provided, the data will be mapped, otherwise it uses the DefaultMapper.
type Config struct {
	name        string
	EXPTypeName string `mapstructure:"type_name"`
	EXPEndpoint string `mapstructure:"endpoint"`
	EXPInterval string `mapstructure:"interval"`
	EXPTimeout  string `mapstructure:"timeout"`
	EXPBackoff  int    `mapstructure:"backoff"`
	MapFile     string `mapstructure:"map_file"`

	log      internal.FieldLogger
	interval time.Duration
	timeout  time.Duration
	mapper   datatype.Mapper
}

// Conf func is used for initializing a Config object.
type Conf func(*Config) error

// NewConfig returns an instance of the expvar reader
// func NewConfig(log internal.FieldLogger, name, typeName string, endpoint, routepath string, interval, timeout time.Duration, backoff int, mapFile string) (*Config, error) {
func NewConfig(conf ...Conf) (*Config, error) {
	obj := new(Config)
	for _, c := range conf {
		err := c(obj)
		if err != nil {
			return nil, err
		}
	}

	if obj.name == "" {
		return nil, reader.ErrEmptyName
	}
	if obj.EXPEndpoint == "" {
		return nil, reader.ErrEmptyEndpoint
	}
	if obj.EXPTypeName == "" {
		return nil, reader.ErrEmptyTypeName
	}
	if obj.mapper == nil {
		obj.mapper = datatype.DefaultMapper()
	}
	return obj, nil
}

// NewInstance returns an instance of the expvar reader.
func (c *Config) NewInstance() (reader.DataReader, error) {
	return New(
		reader.WithLogger(c.Logger()),
		reader.WithEndpoint(c.Endpoint()),
		reader.WithMapper(c.mapper),
		reader.WithName(c.Name()),
		reader.WithTypeName(c.EXPTypeName),
		reader.WithInterval(c.Interval()),
		reader.WithTimeout(c.Timeout()),
		reader.WithBackoff(c.Backoff()),
	)
}

// Name returns name
func (c *Config) Name() string { return c.name }

// TypeName returns type name
func (c *Config) TypeName() string { return c.EXPTypeName }

// Endpoint returns endpoint
func (c *Config) Endpoint() string { return c.EXPEndpoint }

// Interval returns interval
func (c *Config) Interval() time.Duration { return c.interval }

// Timeout returns timeout
func (c *Config) Timeout() time.Duration { return c.timeout }

// Logger returns logger
func (c *Config) Logger() internal.FieldLogger { return c.log }

// Backoff returns backoff
func (c *Config) Backoff() int { return c.EXPBackoff }

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
			return reader.ErrEmptyName
		}
		c.name = name
		return nil
	}
}

// WithTypeName produces an error if typeName is empty.
func WithTypeName(typeName string) Conf {
	return func(c *Config) error {
		if typeName == "" {
			return fmt.Errorf("invalid typeName: %s", typeName)
		}
		c.EXPTypeName = typeName
		return nil
	}
}

// WithEndpoint produces an error if endpoint is empty.
func WithEndpoint(endpoint string) Conf {
	return func(c *Config) error {
		if endpoint == "" {
			return reader.ErrEmptyEndpoint
		}
		endpoint, err := internal.SanitiseURL(endpoint)
		if err != nil {
			return reader.ErrInvalidEndpoint(endpoint)
		}
		c.EXPEndpoint = endpoint
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
		if v == nil {
			return errors.New("no config file")
		}
		err := v.UnmarshalKey(key, &c)
		if err != nil || v.AllKeys() == nil {
			return errors.Wrap(err, "decoding config")
		}

		var interval, timeout time.Duration
		if interval, err = time.ParseDuration(c.EXPInterval); err != nil {
			return errors.Wrapf(err, "parse interval (%v)", c.EXPInterval)
		}
		c.interval = interval

		if timeout, err = time.ParseDuration(c.EXPTimeout); err != nil {
			return errors.Wrapf(err, "parse timeout (%v)", c.EXPTimeout)
		}
		// if c.SelfTypeName == "" {
		// 	return nil, fmt.Errorf("type_name cannot be empty: %s", c.SelfTypeName)
		// }

		c.timeout = timeout
		c.name = name

		if c.MapFile != "" {
			WithMapFile(c.MapFile)
		}

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
		c.EXPBackoff = backoff
		return nil
	}
}

// WithInterval doesn't produce any errors.
func WithInterval(interval time.Duration) Conf {
	return func(c *Config) error {
		c.interval = interval
		return nil
	}
}

// WithMapFile returns any errors on reading the file.
// If the mapFile is empty, it does nothing and returns nil.
func WithMapFile(mapFile string) Conf {
	return func(c *Config) error {
		if mapFile != "" {
			extension := filepath.Ext(mapFile)
			filename := mapFile[0 : len(mapFile)-len(extension)]
			v := viper.New()
			v.SetConfigName(filename)
			v.SetConfigType("yaml")
			v.AddConfigPath(".")
			err := v.ReadInConfig()
			if err != nil {
				return err
			}
			c.mapper = datatype.MapsFromViper(v)
		}

		return nil
	}
}
