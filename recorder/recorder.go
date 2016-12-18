// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder

import (
    "context"
    "time"

    "github.com/arsham/expvastic/datatype"
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

// RecordJob is sent with a context and a payload to be recorded
// If the TypeName and IndexName are different than the previous one, the recorder should use the new ones
type RecordJob struct {
    Ctx       context.Context
    Payload   datatype.DataContainer
    IndexName string
    TypeName  string
    Time      time.Time // Is used for timeseries data
    Err       chan error
}
