// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/antonholmquist/jason"
	"github.com/pkg/errors"
)

// TimeStampFormat specifies the format that all timestamps should be formatted with.
var TimeStampFormat = "2006-01-02T15:04:05.999999-07:00"

// Container satisfies the DataContainer
type Container struct {
	sync.RWMutex
	list []DataType
}

// New returns a new container and populates it with the given list.
func New(list []DataType) *Container {
	return &Container{list: list}
}

// List returns the data.
func (c *Container) List() []DataType {
	c.RLock()
	defer c.RUnlock()
	return c.list
}

// Len returns the length of the data.
func (c *Container) Len() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.list)
}

// Add adds d to the list. You can pass it as many items you need to.
func (c *Container) Add(d ...DataType) {
	c.Lock()
	c.list = append(c.list, d...)
	c.Unlock()
}

// Generate prepends a timestamp pair and value to the list, and generates
// a json object suitable for recording into a document store.
func (c *Container) Generate(p io.Writer, timestamp time.Time) (int, error) {
	ts := fmt.Sprintf(`"@timestamp":"%s"`, timestamp.Format(TimeStampFormat))
	l := new(bytes.Buffer)
	for _, v := range c.List() {
		l.Write([]byte(","))
		_, err := v.Write(l)
		if err != nil {
			return 0, errors.Wrap(err, "writing item")
		}
	}
	ls := l.Bytes()
	return p.Write([]byte(fmt.Sprintf("{%s%s}", ts, ls)))
}

// JobResultDataTypes generates a list of DataType and puts them inside the DataContainer.
// It returns errors if unmarshaling is unsuccessful or ErrUnidentifiedJason when the container
// ends up empty.
func JobResultDataTypes(b []byte, mapper Mapper) (DataContainer, error) {
	obj, err := jason.NewObjectFromBytes(b)
	if err != nil {
		return nil, err
	}
	payload := mapper.Values("", obj.Map())

	if len(payload) == 0 {
		expUnidentifiedJSON.Add(1)
		return nil, ErrUnidentifiedJason
	}
	return New(payload), nil
}
