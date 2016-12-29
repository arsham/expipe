// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package self

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/reader"
	"github.com/spf13/viper"
)

// Config describes how expvastic's own app is setup
// IMPORTANT NOTE: This file was copied to elasticsearch's config. Please create tests for that one if this API changes.

// Config holds the necessary configuration for setting up an self reading facility.
type Config struct {
	name         string
	SelfTypeName string `mapstructure:"type_name"`
	SelfInterval string `mapstructure:"interval"`
	mapper       datatype.Mapper

	interval time.Duration
	log      logrus.FieldLogger
}

// FromViper constructs the necessary configuration for bootstrapping the expvar reader
func FromViper(v *viper.Viper, log logrus.FieldLogger, name, key string) (*Config, error) {
	var (
		c     Config
		inter time.Duration
	)
	err := v.UnmarshalKey(key, &c)
	if err != nil {
		return nil, fmt.Errorf("decoding config: %s", err)
	}
	if inter, err = time.ParseDuration(c.SelfInterval); err != nil {
		return nil, fmt.Errorf("parse interval (%v): %s", c.SelfInterval, err)
	}

	if c.SelfTypeName == "" {
		return nil, fmt.Errorf("type_name cannot be empty: %s", c.SelfTypeName)
	}
	c.mapper = datatype.DefaultMapper()
	c.interval = inter
	c.log = log
	c.name = name
	return &c, nil
}

// NewInstance instantiates a SelfReader
func (c *Config) NewInstance(ctx context.Context, jobChan chan context.Context, resultChan chan *reader.ReadJobResult, errorChan chan<- communication.ErrorMessage) (reader.DataReader, error) {
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	l.Close()
	url := "http://" + l.Addr().String() + "/debug/vars"
	go http.ListenAndServe(l.Addr().String(), nil)
	c.log.Debugf("running self expvar on %s", l.Addr().String())
	return NewSelfReader(c.log, url, c.mapper, jobChan, resultChan, errorChan, c.name, c.TypeName(), c.interval, time.Second)
}

// Name returns the name
func (c *Config) Name() string { return c.name }

// TypeName returns the typename
func (c *Config) TypeName() string { return c.SelfTypeName }

// Endpoint returns the endpoint
func (c *Config) Endpoint() string { return "" }

// RoutePath returns the routepath
func (c *Config) RoutePath() string { return "" }

// Interval returns the interval
func (c *Config) Interval() time.Duration { return c.interval }

// Timeout returns the timeout
func (c *Config) Timeout() time.Duration { return 0 }

// Logger returns the logger
func (c *Config) Logger() logrus.FieldLogger { return c.log }

// Backoff returns the backoff
func (c *Config) Backoff() int { return 0 }
