// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
    "context"
    "io"
    "time"
)

// TargetReader recieves job requests to read from the target, and returns its success results with in a JobResult channel.
type TargetReader interface {
    // The engine will send a signal to this channel to inform the reader when to read from the target.
    // The engine never blocks if the reader is not able to process the requests. It is the reader's job to provide large enough
    // channel, otherwise it will cause goroutine leakage.
    // The context might be canceled depending how the user sets the timeouts.
    JobChan() chan context.Context

    // The engine runs the reading job in another goroutine. Therefore it is the reader's job not to send send a lot of results back
    // to the engine, otherwise it might cause crash on readers and the application itself.
    ResultChan() chan *ReadJobResult

    // The reader's loop should be inside a goroutine, and return a channel.
    // This channel should be closed one it's work is finished and wants to quit.
    Start() chan struct{}

    Name() string
}

// ReadJob is sent with a context and a channel to read the errors
type ReadJob struct {
    Ctx context.Context
    Err chan error
}

// ReadJobResult is constructed everytime a new record is fetched.
// The time is set after the request was successfully read
type ReadJobResult struct {
    Time time.Time
    Res  io.ReadCloser
    Err  error
}
