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

// TimeStampFormat specifies the format that all timestamps should be formatted with.
var TimeStampFormat = "2006-01-02T15:04:05.999999-07:00"

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
	ts := fmt.Sprintf(`"@timestamp":"%s"`, timestamp.Format(TimeStampFormat))
	l := make([][]byte, c.Len()+1)
	l[0] = []byte(ts)
	for i, v := range c.List() {
		l[i+1] = v.Bytes()
	}
	return []byte(fmt.Sprintf("{%s}", bytes.Join(l, []byte(","))))
}
