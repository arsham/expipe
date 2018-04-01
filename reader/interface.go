// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package reader contains logic for reading from a provider. Any types that
// implements the DataReader interface can be used in this system. The Result
// should provide a byte slice that is JSON unmarshallable, otherwise the data
// will be rejected.
//
// Important Notes
//
// When the token's context is cancelled, the reader should finish its job and
// return. The Time should be set when the data is read from the endpoint,
// otherwise it will lose its meaning. The engine will issue jobs based on the
// Interval, which is set in the configuration file.
package reader

import (
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/tools/token"
)

// DataReader receives job requests to read from the target. It returns
// an error if the data cannot be read or the connection is refused.
//
// Notes
//
// Readers should not intercept the engine's decision on the TypeName, unless
// they have a valid reason.
// Ping should ping the endpoint and return nil if was successful. The Engine
// will not launch the reader if the ping result is an error.
// When the context is timed-out or cancelled, Read should return.
// The Engine uses returned object from Mapper to present the data to recorders.
// TypeName is usually the application name and is set by the user in the
// configuration file.
type DataReader interface {
	Name() string
	Read(*token.Context) (*Result, error)
	Ping() error
	Mapper() datatype.Mapper
	TypeName() string
	Timeout() time.Duration
	Interval() time.Duration
	Endpoint() string
	Backoff() int
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
