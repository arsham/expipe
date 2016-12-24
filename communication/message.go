// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package communication contains necessary logic for passing messages
// and returning errors. The Engine will issue a ReadJob with a context
// and sends it to the readers. The NewReadJob function injects a unique ID
// into the context and returns it.
// All readers/recorders should use this JobID for returning errors and logging.
package communication

import (
    "context"

    uuid "github.com/satori/go.uuid"
)

var messageID = contextKey("message-id")

type contextKey string

func (c contextKey) String() string { return "Message ID: " + string(c) }

// JobID is a unique ID. Only the Engine issues this ID and you should pass it along as you recieve.
type JobID uuid.UUID

func (j JobID) String() string { return uuid.UUID(j).String() }

// NewJobID returns a new unique ID
func NewJobID() JobID {
    return JobID(uuid.NewV4())
}

// ReadJob is a package we send to readers to do their work.
type ReadJob struct {
    ctx context.Context
    id  JobID
}

// NewReadJob constructs a ReadJob with the provided context
func NewReadJob(ctx context.Context) context.Context {
    return context.WithValue(ctx, messageID, NewJobID())
}

// ID returns the id of the message
func (r *ReadJob) ID() JobID { return r.id }

// Context returns the context of the message
func (r *ReadJob) Context() context.Context { return r.ctx }
func (r *ReadJob) String() string           { return r.id.String() }

// JobValue returns the value recorder in the context
func JobValue(ctx context.Context) JobID {
    return ctx.Value(messageID).(JobID)
}

// An ErrorMessage is sent when an error occures.
type ErrorMessage struct {
    // The ID comes from the issued job.
    ID JobID
    // Name is the name of the instance, which is returned by its Name() method
    Name string
    Err  error
}

func (e *ErrorMessage) Error() string { return e.Err.Error() }
