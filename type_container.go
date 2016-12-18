// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
    "fmt"
    "strings"
    "sync"
    "time"
)

// DataContainer is an interface for holding a list of DataType.
// I'm aware of the container/list package, which is awesome, but I needed to guard this with a mutex.
// To iterate over this contaner:
//  for i := 0; i < d.Len(); i++ {
//      item := d.Get(i)
//  }
type DataContainer interface {
    // Returns the list. You should not update this list as it is a shared list and anyone can read from it.
    // If you append to this list, there is a chance you are not refering to the same underlying array in memory.
    List() []DataType
    Len() int

    String(timestamp time.Time) string
    // Returns the Err value
    Error() error
}

// Container satisfies the DataContainer and error interfaces
type Container struct {
    mu   sync.RWMutex
    list []DataType
    Err  error
}

// NewContainer returns a new container
func NewContainer(list []DataType) *Container {
    return &Container{list: list}
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

// TODO: Use JSON encoder instead
func (c *Container) String(timestamp time.Time) string {
    ts := fmt.Sprintf(`"@timestamp":"%s"`, timestamp.Format("2006-01-02T15:04:05.999999-07:00"))
    l := make([]string, c.Len()+1)
    l[0] = ts

    for i, v := range c.List() {
        l[i+1] = v.String()
    }
    return fmt.Sprintf("{%s}", strings.Join(l, ","))
}
