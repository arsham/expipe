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
	name       string
	TypeName_  string `mapstructure:"type_name"`
	Endpoint_  string `mapstructure:"endpoint"`
	RoutePath_ string `mapstructure:"routepath"`
	Interval_  string `mapstructure:"interval"`
	Timeout_   string `mapstructure:"timeout"`
	Backoff_   int    `mapstructure:"backoff"`
	MapFile    string `mapstructure:"map_file"`

	log      logrus.FieldLogger
	interval time.Duration
	timeout  time.Duration
	mapper   datatype.Mapper
}

func NewConfig(
	log logrus.FieldLogger,
	name, typeName string,
	endpoint, routepath string,
	interval, timeout time.Duration,
	backoff int,
	mapFile string,
) (*Config, error) {
	c := &Config{
		name:       name,
		TypeName_:  typeName,
		Endpoint_:  endpoint,
		RoutePath_: routepath,
		timeout:    timeout,
		interval:   interval,
		log:        log,
		Backoff_:   backoff,
		MapFile:    mapFile,
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
		return nil, fmt.Errorf("decodeing config: %s", err)
	}
	c.name = name
	c.log = log

	if interval, err = time.ParseDuration(c.Interval_); err != nil {
		return nil, fmt.Errorf("parse interval (%v): %s", c.Interval_, err)
	}
	c.interval = interval

	if timeout, err = time.ParseDuration(c.Timeout_); err != nil {
		return nil, fmt.Errorf("parse timeout: %s", err)
	}
	c.timeout = timeout

	return withConfig(&c)
}

func withConfig(c *Config) (*Config, error) {
	if c.name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	if c.Endpoint_ == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}

	url, err := lib.SanitiseURL(c.Endpoint_)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %d", c.Endpoint_)
	}
	c.Endpoint_ = url

	if c.Backoff_ <= 5 {
		return nil, fmt.Errorf("back off should be at least 5: %d", c.Backoff_)
	}

	if c.TypeName_ == "" {
		return nil, fmt.Errorf("type_name cannot be empty")
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

func (c *Config) NewInstance(ctx context.Context, jobChan chan context.Context, resultChan chan *reader.ReadJobResult) (reader.DataReader, error) {
	endpoint, err := url.Parse(c.Endpoint())
	if err != nil {
		return nil, err
	}
	endpoint.Path = path.Join(endpoint.Path, c.RoutePath())
	ctxReader := reader.NewCtxReader(endpoint.String())
	return NewExpvarReader(c.log, ctxReader, c.mapper, jobChan, resultChan, c.name, c.TypeName_, c.interval, c.timeout)
}

func (c *Config) Name() string               { return c.name }
func (c *Config) TypeName() string           { return c.TypeName_ }
func (c *Config) Endpoint() string           { return c.Endpoint_ }
func (c *Config) RoutePath() string          { return c.RoutePath_ }
func (c *Config) Interval() time.Duration    { return c.interval }
func (c *Config) Timeout() time.Duration     { return c.timeout }
func (c *Config) Logger() logrus.FieldLogger { return c.log }
func (c *Config) Backoff() int               { return c.Backoff_ }
