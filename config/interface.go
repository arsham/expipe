// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package config reads the configurations from a yaml file and produces necessary
// configuration for instantiating readers and recorders.
package config

import (
	"context"
	"time"

	"github.com/arsham/expvastic/internal"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/recorder"
)

// Conf will return necessary information for setting up readers and recorders.
type Conf interface {
	// Endpoint should be a valid URL.
	Endpoint() string

	// Timeout is how long to wait for a response.
	Timeout() time.Duration

	// Backoff returns the amount of retries after the endpoint is rejected
	// the request or not responsive.
	Backoff() int

	// Logger returns an instance of the logger.
	Logger() internal.FieldLogger
}

// ReaderConf defines a behaviour of a reader.
type ReaderConf interface {
	Conf

	// Interval is used to signal the reader to fetch the next read.
	Interval() time.Duration

	// NewInstance should return an initialised Reader instance.
	NewInstance(ctx context.Context) (reader.DataReader, error)

	// TypeName is usually the application name.
	// Readers should not intercept the engine's decision on setting
	// the name, unless they have a valid reason.
	TypeName() string
}

// RecorderConf defines a behaviour that requires the recorders to have the concept
// of Index and Type. If any of these are not applicable, just return an empty string.
type RecorderConf interface {
	Conf

	// NewInstance should return an initialised Recorder instance.
	NewInstance(ctx context.Context) (recorder.DataRecorder, error)

	// IndexName comes from the configuration, but the engine might take over.
	// Recorders should not intercept the engine's decision on setting
	// the name, unless they have a valid reason.
	IndexName() string
}
