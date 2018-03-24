// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
	"io"
	"time"

	"github.com/antonholmquist/jason"
)

// DataType represents a single paired data. The key of the json value
// is mapped to Key, and the value is to Value.
// Write includes both Key and Value.
// THe Equal method comparison of the value is not ordered.
type DataType interface {
	Write(p io.Writer) (n int, err error)
	Equal(other DataType) bool
}

// DataContainer is an interface for holding a list of DataType.
// I'm aware of the container/list package, which is awesome, but I needed
// a simple interface to do this job. You should not update the list returning
// from List as it is a shared list and anyone can read from it. If you append
// to this list, there is a  chance you are not referring to the same underlying
// array in memory. Generate writes to w applying the timestamp. It returns the
// amount of bytes it's written and an error if any of its contents report.
type DataContainer interface {
	List() []DataType
	Len() int
	Generate(p io.Writer, timestamp time.Time) (n int, err error)
}

// Mapper generates DataTypes based on the given name/value inputs.
// Values closes the channel once all input has been exhausted.
// You should always copy the mapper if you are using it concurrently.
type Mapper interface {
	Values(prefix string, values map[string]*jason.Value) []DataType
	Copy() Mapper
}
