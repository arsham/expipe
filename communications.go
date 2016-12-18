// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
    "context"
    "io"
    "time"
)

// DataRecorder in an interface for shipping data to a repository.
// The repository should have the concept of index/database and type/table abstractions. See ElasticSearch for more information.
// Recorder should send nil to Err channel if no error occurs.
type DataRecorder interface {
    // Reader should not block when RecordJob is sent to this channel.
    PayloadChan() chan *RecordJob

    // The recorder's loop should be inside a goroutine, and return a channel.
    // This channel should be closed one it's work is finished and wants to quit.
    Start() chan struct{}
}

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
}

// RecordJob is sent with a context and a payload to be recorded
// If the TypeName and IndexName are different than the previous one, the recorder should use the new ones
type RecordJob struct {
    Ctx       context.Context
    Payload   DataContainer
    IndexName string
    TypeName  string
    Time      time.Time // Is used for timeseries data
    Err       chan error
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
