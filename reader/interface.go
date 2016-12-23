// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package reader contains logic for reading from a provider. Any objects that implements the DataReader interface
// can be used in this system. The job should provide an io.ReadCloser and should produce a JSON object, othewise
// the data will be rejected.
//
// The data stream SHOULD not be closed. The engine WILL close it upon reading its contents.
//
// Readers should ping their endpoint upon creation to make sure they can read from. Otherwise they should return
// an error indicating they cannot start.
//
// When the context is canceled, the reader should finish its job and return. The Time should be set when the data is
// read from the endpoint, otherwise it will lose its meaning. The engine will issue jobs based on the Interval, which
// is set in the configuration file.
package reader

import (
	"context"
	"io"
	"time"
)

// InterTimer is required by the Engine so it can read the intervals and timeouts.
type InterTimer interface {
	Timeout() time.Duration  // is used for timing out reading from the endpoint.
	Interval() time.Duration // the engine will issue jobs based on this interval.
}

// DataReader recieves job requests to read from the target, and sends its success
// through the ResultChan channel.
type DataReader interface {
	InterTimer

	// The engine will send a signal to this channel to inform the reader when it is time to read
	// from the target.
	// The engine never blocks if the reader is not able to process the requests.
	// This channel will be provided by the Engine.
	// The context might be canceled depending how the user sets the timeouts.
	JobChan() chan context.Context

	// The engine runs the reading job in another goroutine. The engine will provide this channel, however
	// the reader should not send send a lot of data back to the engine, otherwise it might cause crash
	// on readers and the application itself.
	ResultChan() chan *ReadJobResult

	// The reader's loop should be inside a goroutine, and return a done channel.
	// This channel should be closed once its work is finished and the reader wants to quit.
	// When the context is timedout or canceled, the reader should return.
	Start(ctx context.Context) <-chan struct{}

	// TypeName is usually the application name and is set by the user in the configuration file.
	// Recorders should not intercept the engine for its decision, unless they have a
	// valid reason.
	TypeName() string

	// Name should return the representation string for this reader. Choose a very simple and unique name.
	Name() string
}

// ReadJob is sent with a context and a channel to read the errors back.
type ReadJob struct {
	Ctx context.Context

	// The reader should always send an error messge back, even if there is no errors otherwise it will
	// cause goroutine leakage. In case there were no error, just send nil.
	Err chan error
}

// ReadJobResult is constructed everytime a new record is fetched.
// The time is set after the request was successfully read.
type ReadJobResult struct {
	Time     time.Time
	TypeName string
	Res      io.ReadCloser
	Err      error
}
