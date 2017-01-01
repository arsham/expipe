// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvar

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	"github.com/spf13/viper"
)

// Config describes how expvar read is setup
// IMPORTANT NOTE: This file was copied to elasticsearch's config. Please create tests for that one if this API changes.

// Config holds the necessary configuration for setting up an expvar reader endpoint.
// If MapFile is provided, the data will be mapped, otherwise it uses the defaults.
type Config struct {
	name         string
	EXPTypeName  string `mapstructure:"type_name"`
	EXPEndpoint  string `mapstructure:"endpoint"`
	EXPRoutePath string `mapstructure:"routepath"`
	EXPInterval  string `mapstructure:"interval"`
	EXPTimeout   string `mapstructure:"timeout"`
	EXPBackoff   int    `mapstructure:"backoff"`
	MapFile      string `mapstructure:"map_file"`

	log      logrus.FieldLogger
	interval time.Duration
	timeout  time.Duration
	mapper   datatype.Mapper
}

// NewConfig returns an instance of the expvar reader
func NewConfig(
	log logrus.FieldLogger,
	name, typeName string,
	endpoint, routepath string,
	interval, timeout time.Duration,
	backoff int,
	mapFile string,
) (*Config, error) {
	c := &Config{
		name:         name,
		EXPTypeName:  typeName,
		EXPEndpoint:  endpoint,
		EXPRoutePath: routepath,
		timeout:      timeout,
		interval:     interval,
		log:          log,
		EXPBackoff:   backoff,
		MapFile:      mapFile,
	}
	return withConfig(c)
}

// FromViper constructs the necessary configuration for bootstrapping the expvar reader
func FromViper(v *viper.Viper, log logrus.FieldLogger, name, key string) (*Config, error) {
	var (
		c                 Config
		interval, timeout time.Duration
	)
	err := v.UnmarshalKey(key, &c)
	if err != nil {
		return nil, fmt.Errorf("decoding config: %s", err)
	}
	c.name = name
	c.log = log

	if interval, err = time.ParseDuration(c.EXPInterval); err != nil {
		return nil, fmt.Errorf("parse interval (%v): %s", c.EXPInterval, err)
	}
	c.interval = interval

	if timeout, err = time.ParseDuration(c.EXPTimeout); err != nil {
		return nil, fmt.Errorf("parse timeout: %s", err)
	}
	c.timeout = timeout

	return withConfig(&c)
}

func withConfig(c *Config) (*Config, error) {
	// TODO: these checks are also in the reader, remove them
	if c.name == "" {
		return nil, reader.ErrEmptyName
	}

	if c.EXPEndpoint == "" {
		return nil, reader.ErrEmptyEndpoint
	}

	url, err := lib.SanitiseURL(c.EXPEndpoint)
	if err != nil {
		return nil, reader.ErrInvalidEndpoint(c.EXPEndpoint)
	}
	c.EXPEndpoint = url

	if c.EXPTypeName == "" {
		return nil, reader.ErrEmptyTypeName
	}

	if c.MapFile != "" {
		extension := filepath.Ext(c.MapFile)
		filename := c.MapFile[0 : len(c.MapFile)-len(extension)]
		v := viper.New()
		v.SetConfigName(filename)
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		err := v.ReadInConfig()
		if err != nil {
			c.log.Warnf("Maps file not found or contains error, loading the default settings: %s", err)
		} else {
			c.mapper = datatype.MapsFromViper(v)
		}
	}

	if c.mapper == nil {
		c.mapper = datatype.DefaultMapper()
	}

	return c, nil
}

// NewInstance returns an instance of the expvar reader
func (c *Config) NewInstance(ctx context.Context) (reader.DataReader, error) {
	endpoint, err := url.Parse(c.Endpoint())
	if err != nil {
		return nil, err
	}
	endpoint.Path = path.Join(endpoint.Path, c.RoutePath())
	return NewExpvarReader(c.Logger(), endpoint.String(), c.mapper, c.Name(), c.EXPTypeName, c.Interval(), c.Timeout(), c.Backoff())
}

// Name returns name
func (c *Config) Name() string { return c.name }

// TypeName returns type name
func (c *Config) TypeName() string { return c.EXPTypeName }

// Endpoint returns endpoint
func (c *Config) Endpoint() string { return c.EXPEndpoint }

// RoutePath returns routepath
func (c *Config) RoutePath() string { return c.EXPRoutePath }

// Interval returns interval
func (c *Config) Interval() time.Duration { return c.interval }

// Timeout returns timeout
func (c *Config) Timeout() time.Duration { return c.timeout }

// Logger returns logger
func (c *Config) Logger() logrus.FieldLogger { return c.log }

// Backoff returns backoff
func (c *Config) Backoff() int { return c.EXPBackoff }
