// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package reader contains logic for reading from a provider. Any types that implements the DataReader interface
// can be used in this system. The Result should provide a byte slice that is JSON unmarshallable, otherwise
// the data will be rejected.
//
// Important Notes
//
// When the token's context is cancelled, the reader should finish its job and return. The Time should be set when the data is
// read from the endpoint, otherwise it will lose its meaning. The engine will issue jobs based on the Interval, which
// is set in the configuration file.
package reader

import (
	"time"

	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/token"
)

// DataReader receives job requests to read from the target. It returns
// an error if the data cannot be read or the connection is refused.
//
// Notes
//
// Readers should not intercept the engine's decision on the TypeName,
// unless they have a valid reason.
type DataReader interface {
	// Name should return the representation string for this reader.
	// Choose a very simple and unique name.
	Name() string

	// Ping should ping the endpoint and return nil if was successful.
	// The Engine will not launch the reader if the ping result is an error.
	Ping() error

	// When the context is timed-out or cancelled, the reader should return.
	Read(*token.Context) (*Result, error)

	// Mapper should return an instance of the datatype mapper.
	// Engine uses this object to present the data to recorders.
	Mapper() datatype.Mapper

	// TypeName is usually the application name and is set by the user in
	// the configuration file.
	TypeName() string

	// Timeout is required by the Engine so it can read the time-outs.
	Timeout() time.Duration

	// Interval is required by the Engine so it can read the intervals.
	Interval() time.Duration
}

// Result is constructed every time a new data is fetched.
type Result struct {
	// ID is the job ID given by the Engine.
	// This ID should not be changed until it is recorded.
	ID token.ID

	//Time is set after the request was successfully read.
	Time time.Time

	// TypeName comes from the configuration, but the Engine might decide
	// to change it.
	TypeName string

	// Content should be json unmarshallable, otherwise the job will be dropped.
	Content []byte

	// Mapper is the mapper set in the reader.
	Mapper datatype.Mapper
}
