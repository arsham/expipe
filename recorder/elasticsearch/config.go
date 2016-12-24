// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package elasticsearch

import (
    "context"
    "fmt"
    "time"

    "github.com/Sirupsen/logrus"
    "github.com/arsham/expvastic/lib"
    "github.com/arsham/expvastic/recorder"
    "github.com/spf13/viper"
)

// Config holds the necessary configuration for setting up an elasticsearch reader endpoint.
type Config struct {
    name       string
    Endpoint_  string `mapstructure:"endpoint"`
    Timeout_   string `mapstructure:"timeout"`
    Backoff_   int    `mapstructure:"backoff"`
    IndexName_ string `mapstructure:"index_name"`

    logger  logrus.FieldLogger
    timeout time.Duration
}

func NewConfig(log logrus.FieldLogger, name string, endpoint string, timeout time.Duration, backoff int, indexName string) (*Config, error) {
    c := &Config{
        name:       name,
        Endpoint_:  endpoint,
        timeout:    timeout,
        Backoff_:   backoff,
        IndexName_: indexName,
        logger:     log,
    }
    return withConfig(c)
}

// FromViper constructs the necessary configuration for bootstrapping the elasticsearch reader
func FromViper(v *viper.Viper, log logrus.FieldLogger, name, key string) (*Config, error) {
    var (
        c       Config
        timeout time.Duration
    )
    err := v.UnmarshalKey(key, &c)
    if err != nil {
        return nil, fmt.Errorf("decodeing config: %s", err)
    }

    c.name = name
    c.logger = log

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
    endpoint, err := lib.SanitiseURL(c.Endpoint_)
    if err != nil {
        return nil, fmt.Errorf("invalid endpoint: %d", c.Endpoint_)
    }
    c.Endpoint_ = endpoint

    if c.IndexName_ == "" {
        return nil, fmt.Errorf("index_name cannot be empty")
    }

    if c.Backoff_ <= 5 {
        return nil, fmt.Errorf("back off should be at least 5: %d", c.Backoff_)
    }
    return c, nil
}

func (c *Config) NewInstance(ctx context.Context, payloadChan chan *recorder.RecordJob) (recorder.DataRecorder, error) {
    return NewRecorder(ctx, c.logger, payloadChan, c.name, c.Endpoint(), c.IndexName(), c.timeout)
}
func (c *Config) Name() string               { return c.name }
func (c *Config) IndexName() string          { return c.IndexName_ }
func (c *Config) Endpoint() string           { return c.Endpoint_ }
func (c *Config) RoutePath() string          { return "" }
func (c *Config) Timeout() time.Duration     { return c.timeout }
func (c *Config) Logger() logrus.FieldLogger { return c.logger }
func (c *Config) Backoff() int               { return c.Backoff_ }
