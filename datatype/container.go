// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
	"bytes"
	"fmt"
	"sync"
	"time"
)

// DataContainer is an interface for holding a list of DataType.
// I'm aware of the container/list package, which is awesome, but I needed
// a simple interface to do this job.
type DataContainer interface {
	// List returns the list. You should not update this list as it is a shared
	// list and anyone can read from it. If you append to this list, there is a
	// chance you are not referring to the same underlying array in memory.
	List() []DataType

	// Len returns the length of the container.
	Len() int

	// Bytes returns the []byte representation of the container by collecting
	// all []byte values of its contents.
	Bytes(timestamp time.Time) []byte

	// Returns the Err value.
	Error() error
}

// Container satisfies the DataContainer and error interfaces.
type Container struct {
	// Err value is set during container creation.
	Err  error
	mu   sync.RWMutex
	list []DataType
}

// New returns a new container and populates it with the given list.
func New(list []DataType) *Container {
	return &Container{list: list}
}

// List returns the data.
// The error is not provided here, please check the Err value.
func (c *Container) List() []DataType {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.list
}

// Len returns the length of the data.
func (c *Container) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.list)
}

// Add adds d to the list. You can pass it as many items you need to.
func (c *Container) Add(d ...DataType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.list = append(c.list, d...)
}

// Error returns the error message.
func (c *Container) Error() error {
	return c.Err
}

// Bytes prepends a timestamp pair and value to the list, and generates
// a json object suitable for recording into a document store.
func (c *Container) Bytes(timestamp time.Time) []byte {
	ts := fmt.Sprintf(`"@timestamp":"%s"`, timestamp.Format("2006-01-02T15:04:05.999999-07:00"))
	l := make([][]byte, c.Len()+1)
	l[0] = []byte(ts)
	for i, v := range c.List() {
		l[i+1] = v.Bytes()
	}
	return []byte(fmt.Sprintf("{%s}", bytes.Join(l, []byte(","))))
}
