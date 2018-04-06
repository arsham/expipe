// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvar

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/tools"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Config holds the necessary configuration for setting up an expvar reader
// endpoint. If MapFile is provided, the data will be mapped, otherwise it uses
// the DefaultMapper.
type Config struct {
	log          tools.FieldLogger
	EXPTypeName  string `mapstructure:"type_name"`
	EXPEndpoint  string `mapstructure:"endpoint"`
	EXPInterval  string `mapstructure:"interval"`
	EXPTimeout   string `mapstructure:"timeout"`
	MapFile      string `mapstructure:"map_file"`
	EXPName      string
	ConfInterval time.Duration
	ConfTimeout  time.Duration
	mapper       datatype.Mapper
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

// Reader implements the ReaderConf interface.
func (c *Config) Reader() (reader.DataReader, error) {
	return New(
		reader.WithLogger(c.Logger()),
		reader.WithEndpoint(c.Endpoint()),
		reader.WithMapper(c.mapper),
		reader.WithName(c.Name()),
		reader.WithTypeName(c.EXPTypeName),
		reader.WithInterval(c.Interval()),
		reader.WithTimeout(c.Timeout()),
	)
}

// Name returns name from the config file.
func (c *Config) Name() string { return c.EXPName }

// TypeName returns type name from the config file.
func (c *Config) TypeName() string { return c.EXPTypeName }

// Endpoint returns endpoint from the config file.
func (c *Config) Endpoint() string { return c.EXPEndpoint }

// Interval returns interval after reading from the config file.
func (c *Config) Interval() time.Duration { return c.ConfInterval }

// Timeout returns timeout after reading from the config file.
func (c *Config) Timeout() time.Duration { return c.ConfTimeout }

// Logger returns logger.
func (c *Config) Logger() tools.FieldLogger { return c.log }

// Mapper returns the mapper assigned to this object.
func (c *Config) Mapper() datatype.Mapper { return c.mapper }

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
		c.ConfInterval = interval

		if timeout, err = time.ParseDuration(c.EXPTimeout); err != nil {
			return errors.Wrapf(err, "parse timeout (%v)", c.EXPTimeout)
		}
		if c.EXPTypeName == "" {
			return fmt.Errorf("type_name cannot be empty: %s", c.EXPTypeName)
		}
		c.ConfTimeout = timeout
		c.EXPName = name
		if c.MapFile != "" {
			WithMapFile(c.MapFile)
		}

		return nil
	}
}

// WithMapFile returns any errors on reading the file. If the mapFile is empty,
// it does nothing and returns nil.
func WithMapFile(mapFile string) Conf {
	return func(c *Config) error {
		if mapFile == "" {
			return nil
		}
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
		return nil
	}
}
