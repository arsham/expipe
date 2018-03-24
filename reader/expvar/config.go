// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvar

import (
	"net/url"
	"path"
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
	name         string
	EXPTypeName  string `mapstructure:"type_name"`
	EXPEndpoint  string `mapstructure:"endpoint"`
	EXPRoutePath string `mapstructure:"routepath"`
	EXPInterval  string `mapstructure:"interval"`
	EXPTimeout   string `mapstructure:"timeout"`
	EXPBackoff   int    `mapstructure:"backoff"`
	MapFile      string `mapstructure:"map_file"`

	log      internal.FieldLogger
	interval time.Duration
	timeout  time.Duration
	mapper   datatype.Mapper
}

// NewConfig returns an instance of the expvar reader
func NewConfig(log internal.FieldLogger, name, typeName string, endpoint, routepath string, interval, timeout time.Duration, backoff int, mapFile string) (*Config, error) {
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

type unmarshaller interface {
	UnmarshalKey(key string, rawVal interface{}) error
}

// FromViper constructs the necessary configuration for bootstrapping the expvar reader.
func FromViper(v unmarshaller, log internal.FieldLogger, name, key string) (*Config, error) {
	var (
		c                 Config
		interval, timeout time.Duration
	)
	err := v.UnmarshalKey(key, &c)
	if err != nil {
		return nil, errors.Wrap(err, "decoding config")
	}
	c.name = name
	c.log = log

	if interval, err = time.ParseDuration(c.EXPInterval); err != nil {
		return nil, errors.Wrapf(err, "parse interval (%v)", c.EXPInterval)
	}
	c.interval = interval

	if timeout, err = time.ParseDuration(c.EXPTimeout); err != nil {
		return nil, errors.Wrap(err, "parse timeout")
	}
	c.timeout = timeout

	return withConfig(&c)
}

func withConfig(c *Config) (*Config, error) {
	// TODO: these checks are also in the reader, remove them (0)
	if c.name == "" {
		return nil, reader.ErrEmptyName
	}

	if c.EXPEndpoint == "" {
		return nil, reader.ErrEmptyEndpoint
	}

	url, err := internal.SanitiseURL(c.EXPEndpoint)
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

// NewInstance returns an instance of the expvar reader.
func (c *Config) NewInstance() (reader.DataReader, error) {
	endpoint, err := url.Parse(c.Endpoint()) // TODO: check this part [not sure]
	if err != nil {
		return nil, errors.Wrap(err, "new config")
	}
	endpoint.Path = path.Join(endpoint.Path, c.RoutePath())
	return New(
		reader.WithLogger(c.Logger()),
		reader.WithEndpoint(endpoint.String()),
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

// RoutePath returns routepath
func (c *Config) RoutePath() string { return c.EXPRoutePath }

// Interval returns interval
func (c *Config) Interval() time.Duration { return c.interval }

// Timeout returns timeout
func (c *Config) Timeout() time.Duration { return c.timeout }

// Logger returns logger
func (c *Config) Logger() internal.FieldLogger { return c.log }

// Backoff returns backoff
func (c *Config) Backoff() int { return c.EXPBackoff }
