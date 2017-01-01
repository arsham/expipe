// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package reader contains logic for reading from a provider. Any objects that implements the DataReader interface
// can be used in this system. The job should provide an io.ReadCloser and should produce a JSON object, otherwise
// the data will be rejected.
//
// The data stream SHOULD not be closed. The engine WILL close it upon reading its contents.
//
// Readers should ping their endpoint upon creation to make sure they can read from. Otherwise they should return
// an error indicating they cannot start.
//
// When the context is cancelled, the reader should finish its job and return. The Time should be set when the data is
// read from the endpoint, otherwise it will lose its meaning. The engine will issue jobs based on the Interval, which
// is set in the configuration file.
package reader

import (
	"context"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
)

// DataReader receives job requests to read from the target, and sends its success
// through the ResultChan channel.
type DataReader interface {
	// Name should return the representation string for this reader. Choose a very simple and unique name.
	Name() string

	// When the context is timed-out or cancelled, the reader should return.
	Read(context.Context) (*ReadJobResult, error)

	// Mapper should return an instance of the datatype mapper.
	// Engine uses this object to present the data to recorders.
	Mapper() datatype.Mapper

	// TypeName is usually the application name and is set by the user in the configuration file.
	// Recorders should not intercept the engine for its decision, unless they have a
	// valid reason.
	TypeName() string

	// Timeout is required by the Engine so it can read the time-outs.
	Timeout() time.Duration

	// Interval is required by the Engine so it can read the intervals.
	Interval() time.Duration
}

// ReadJobResult is constructed every time a new record is fetched.
// The time is set after the request was successfully read.
type ReadJobResult struct {
	ID       communication.JobID
	Time     time.Time
	TypeName string
	Res      []byte
	Mapper   datatype.Mapper //TODO: refactor this out
}
