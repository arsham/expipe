// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package recorder contains logic to record data into a database. The job is
// guaranteed to be json marshallable. Any objects that implements the DataRecorder
// interface can be used in this system.
//
// Recorders should ping their endpoint upon creation to make sure they can access.
// Otherwise they should return an error indicating they cannot start.
//
// When the context is cancelled, the recorder should finish its job and return.
package recorder

import (
	"context"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
)

// DataRecorder receives a payload for shipping data to a repository.
// The repository should have the concept of index/database and type/table
// abstractions. See ElasticSearch for more information.
// Recorders should not change the index name coming in the payload unless
// they have a valid reason. The engine might add a date to this index name
// if the user has specified in the configuration file.
type DataRecorder interface {
	// Name should return the representation string for this recorder.
	// Choose a very simple and unique name.
	Name() string

	// IndexName comes from the configuration, but the engine takes over.
	IndexName() string

	// Timeout is required by the Engine so it can read the time-outs.
	Timeout() time.Duration

	// The recorder should record the RecordJob and report any errors happened.
	// When the context is timed-out or cancelled, the recorder should return
	// with the context's error.
	Record(context.Context, *RecordJob) error
}

// RecordJob is sent with a context and a payload to be recorded.
// If the TypeName and IndexName are different than the previous one, the recorder
// should use the ones engine provides. If any errors occurred, recorders should
// return the error on Read return value.
type RecordJob struct {
	// ID is the job ID generated at the time the payload was generated.
	ID communication.JobID

	// Payload has a Bytes() method for returning the data.
	// It is guaranteed to be json marshallable.
	Payload datatype.DataContainer

	// Time is the recorded time at the time of fetching data by the readers.
	// You should use this value to fetch the content of the payload
	Time time.Time

	// IndexName might be different than the one is set in the recorder.
	// Engine might decide to change it and you have to use the provided one.
	IndexName string

	// TypeName comes from the configuration of readers.
	TypeName string
}
