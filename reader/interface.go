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
	"io"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
)

// InterTimer is required by the Engine so it can read the intervals and time-outs.
type InterTimer interface {
	Timeout() time.Duration  // is used for timing out reading from the endpoint.
	Interval() time.Duration // the engine will issue jobs based on this interval.
}

// DataReader receives job requests to read from the target, and sends its success
// through the ResultChan channel.
type DataReader interface {
	InterTimer

	// The engine will send a signal to this channel to inform the reader when
	// it is time to read from the target.
	// The engine never blocks if the reader is not able to process the requests.
	// This channel will be provided by the Engine.
	// The context might be cancelled depending how the user sets the time-outs.
	// The UUID associated with this job is inside the context. Readers are
	// advised to use this ID and pass them along.
	JobChan() chan context.Context

	// The engine runs the reading job in another goroutine. The engine will provide this channel, however
	// the reader should not send send a lot of data back to the engine, otherwise it might cause crash
	// on readers and the application itself.
	ResultChan() chan *ReadJobResult

	// The reader's loop should be inside a goroutine.
	// This channel should be closed once the worker receives a stop signal
	// and its work is finished. The response to the stop signal should happen
	// otherwise it will hang the Engine around.
	// When the context is timed-out or cancelled, the reader should return.
	Start(ctx context.Context, stop communication.StopChannel)

	// Mapper should return an instance of the datatype mapper.
	// Engine uses this object to present the data to recorders.
	Mapper() datatype.Mapper

	// TypeName is usually the application name and is set by the user in the configuration file.
	// Recorders should not intercept the engine for its decision, unless they have a
	// valid reason.
	TypeName() string

	// Name should return the representation string for this reader. Choose a very simple and unique name.
	Name() string
}

// ReadJobResult is constructed every time a new record is fetched.
// The time is set after the request was successfully read.
type ReadJobResult struct {
	ID       communication.JobID
	Time     time.Time
	TypeName string
	Res      io.ReadCloser
	Mapper   datatype.Mapper //TODO: refactor this out
}
