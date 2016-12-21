// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"context"
	"io"
	"time"
)

// InterTimer is required by the Engine so it can read the intervals and timeouts.
type InterTimer interface {
	Timeout() time.Duration
	Interval() time.Duration
}

// DataReader recieves job requests to read from the target, and sends its success
// through the ResultChan channel.
type DataReader interface {
	InterTimer

	// The engine will send a signal to this channel to inform the reader when to read
	// from the target.
	// The engine never blocks if the reader is not able to process the requests.
	// It is the reader's job to provide a large enough channel, otherwise it will cause goroutine leakage.
	// The context might be canceled depending how the user sets the timeouts.
	JobChan() chan context.Context

	// The engine runs the reading job in another goroutine. Therefore it is the reader's job
	// not to send send a lot of results back to the engine, otherwise it might cause crash
	// on readers and the application itself.
	ResultChan() chan *ReadJobResult

	// The reader's loop should be inside a goroutine, and return a done channel.
	// This channel should be closed once its work is finished and the reader wants to quit.
	// When the context is timedout or canceled, the reader should return.
	Start(ctx context.Context) chan struct{}

	// Name should return the representation string for this reader. Choose a very simple name.
	Name() string
}

// ReadJob is sent with a context and a channel to read the errors back.
type ReadJob struct {
	Ctx context.Context

	// The reader should always send an error messge back, even if there is no errors.
	// In case there were no error, just send nil.
	Err chan error
}

// ReadJobResult is constructed everytime a new record is fetched.
// The time is set after the request was successfully read.
type ReadJobResult struct {
	Time time.Time
	Res  io.ReadCloser
	Err  error
}
