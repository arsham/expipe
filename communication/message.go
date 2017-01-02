// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package communication contains necessary logic for passing messages.
// The Engine will issue a ReadJob with a unique ID and sends it to the reader.
// The NewReadJob function injects a unique ID into the context and returns it.
// All readers/recorders should use this JobID for returning errors and logging.
package communication

import (
	"context"

	uuid "github.com/satori/go.uuid"
)

// messageID will travel with the jobs in a context.
var messageID = contextKey("message-id")

type contextKey string

func (c contextKey) String() string { return "Message ID: " + string(c) }

// JobID is a unique ID. Only the Engine issues this ID and you should pass it along as you receive.
type JobID uuid.UUID

func (j JobID) String() string { return uuid.UUID(j).String() }

// NewJobID returns a new unique ID.
func NewJobID() JobID {
	return JobID(uuid.NewV4())
}

// NewReadJob constructs a ReadJob with the provided context.
func NewReadJob(ctx context.Context) context.Context {
	return context.WithValue(ctx, messageID, NewJobID())
}

// JobValue returns the recorded value in the job.
func JobValue(job context.Context) JobID {
	return job.Value(messageID).(JobID)
}
