// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import "sync"

// DataContainer holds a list of data types
type DataContainer interface {
    List() []DataType
    Len() int
    Error() error
}

// Container satisfies the DataContainer and error interfaces
type Container struct {
    mu   sync.RWMutex
    list []DataType
    Err  error
}

// List returns the data
// The error is not provided here, please check the Err value
func (c *Container) List() []DataType {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.list
}

// Len returns the length of the data
func (c *Container) Len() int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return len(c.list)
}

// Add adds to the list
func (c *Container) Add(d ...DataType) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.list = append(c.list, d...)
}

// Error returns the error
func (c *Container) Error() error {
    return c.Err
}
