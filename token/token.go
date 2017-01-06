// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package token contains necessary logic for passing messages.
// The Engine will issue a token.Context with a unique ID and sends it to the reader.
// The New function injects a unique ID into the context and returns it.
// All readers/recorders should use this ID for returning errors and logging.
package token

import (
	"context"

	uuid "github.com/satori/go.uuid"
)

// tokenID will travel with the jobs in a context.
var tokenID = tokenKey("token-id")

// Context is a context.Context with a token as its value.
type Context struct {
	context.Context
}

// ID returns the recorded value in the c.
func (c *Context) ID() ID {
	return c.Value(tokenID).(ID)
}

type tokenKey string

func (t tokenKey) String() string { return "Token ID: " + string(t) }

// New constructs a ReadJob with the provided context.
func New(ctx context.Context) *Context {
	return &Context{context.WithValue(ctx, tokenID, NewUID())}
}

// ID is a unique ID. Only the Engine issues this ID and you should pass it along as you receive.
type ID uuid.UUID

func (i ID) String() string { return uuid.UUID(i).String() }

// NewUID returns a new unique ID.
func NewUID() ID {
	return ID(uuid.NewV4())
}
