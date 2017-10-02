// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
	"time"

	"github.com/antonholmquist/jason"
)

// DataType represents a single paired data. The key of the json value
// is mapped to Key, and the value is to Value.
type DataType interface {
	// Bytes returns the []byte representation of the value.
	// It includes both Key and Value.
	Bytes() []byte

	// Equal compares the current object to the other returns true if they have
	// equal values. The value comparison is not ordered.
	Equal(other DataType) bool
}

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

// Mapper generates DataTypes based on the given name/value inputs.
type Mapper interface {
	// Values closes the channel once all input has been exhausted.
	Values(prefix string, values map[string]*jason.Value) []DataType

	// Copy returns a new copy of the Mapper.
	// You should always copy the mapper if you are using it concurrently.
	Copy() Mapper
}
