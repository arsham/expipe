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
// The repository should have the concept of type/table. Which is inside index/database abstraction. See ElasticSearch for more information.
// Recorder should send nil to Err channel if no error occurs
type DataRecorder interface {
    // Record(ctx context.Context, typeName string, timestamp time.Time, list DataContainer) error
    PayloadChan() chan *RecordJob
    // Error() error
}

// TargetReader recieves job requests to read from the target, and returns its success results with in a JobResult channel
type TargetReader interface {
    JobChan() chan context.Context
    ResultChan() chan ReadJobResult
    // Error() error
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
