// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package config reads the configurations from a yaml file and produces
// necessary configuration for instantiating readers and recorders.
package config

import (
	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/recorder"
)

// ReaderConf is for configure and returning a Reader instance.
type ReaderConf interface {
	NewInstance() (reader.DataReader, error)
}

// RecorderConf is for configure and returning a Recorder instance.
type RecorderConf interface {
	NewInstance() (recorder.DataRecorder, error)
}
