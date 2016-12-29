// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package datatype contains necessary logic to sanitise a JSON object coming from a reader. This
// package is subjected to change.
package datatype

import (
	"errors"
	"fmt"
	"strings"

	"github.com/arsham/expvastic/lib"
)

const (
	// BYTE ..
	BYTE = 1.0
	// KILOBYTE ..
	KILOBYTE = 1024 * BYTE
	// MEGABYTE ..
	MEGABYTE = 1024 * KILOBYTE
)

// TODO: refactor to use byte slices

// ErrUnidentifiedJason .
var ErrUnidentifiedJason = errors.New("unidentified jason value")

// DataType implements Stringer and Marshal/Unmarshal
type DataType interface {
	fmt.Stringer

	// Equal compares both keys and values and returns true if they are equal
	Equal(DataType) bool
}

// FloatType represents a pair of key values that the value is a float64
type FloatType struct {
	Key   string
	Value float64
}

// String satisfies the Stringer interface
func (f FloatType) String() string {
	return fmt.Sprintf(`"%s":%f`, f.Key, f.Value)
}

// Equal compares both keys and values and returns true if they are equal
// Not implemented
func (f FloatType) Equal(other DataType) bool {
	return false
}

// StringType represents a pair of key values that the value is a string
type StringType struct {
	Key   string
	Value string
}

// String satisfies the Stringer interface
func (s StringType) String() string {
	return fmt.Sprintf(`"%s":"%s"`, s.Key, s.Value)
}

// Equal compares both keys and values and returns true if they are equal
func (s StringType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *StringType:
		return s.Key == o.Key && s.Value == o.Value
	}
	return false
}

// FloatListType represents a pair of key values that the value is a list of floats
type FloatListType struct {
	Key   string
	Value []float64
}

// String satisfies the Stringer interface
func (fl FloatListType) String() string {
	list := make([]string, len(fl.Value))
	for i, v := range fl.Value {
		list[i] = fmt.Sprintf("%f", v)
	}
	return fmt.Sprintf(`"%s":[%s]`, fl.Key, strings.Join(list, ","))
}

// Equal compares both keys and all values and returns true if they are equal.
// The values are checked in an unordered fashion.
func (fl FloatListType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *FloatListType:
		if fl.Key != o.Key {
			return false
		}
		for _, v := range o.Value {
			if !lib.FloatInSlice(v, fl.Value) {
				return false
			}
		}
		return true
	}
	return false
}

// GCListType represents a pair of key values of GC list info
type GCListType struct {
	Key   string
	Value []uint64
}

// String satisfies the Stringer interface
func (flt GCListType) String() string {
	// We are filtering, therefore we don't know the size
	var list []string
	for _, v := range flt.Value {
		if v > 0 {
			list = append(list, fmt.Sprintf("%d", v/1000))
		}
	}
	return fmt.Sprintf(`"%s":[%s]`, flt.Key, strings.Join(list, ","))
}

// Equal is not implemented. You should iterate and check yourself.
// Equal compares both keys and values and returns true if they are equal
func (flt GCListType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *GCListType:
		if flt.Key != o.Key {
			return false
		}
		for _, v := range o.Value {
			if !lib.Uint64InSlice(v, flt.Value) {
				return false
			}
		}
		return true
	}
	return false
}

// ByteType represents a pair of key values in which the value represents bytes
// It converts the value to MB
type ByteType struct {
	Key   string
	Value float64
}

// String satisfies the Stringer interface
func (b ByteType) String() string {
	return fmt.Sprintf(`"%s":%f`, b.Key, b.Value/MEGABYTE)
}

// Equal compares both keys and values and returns true if they are equal
func (b ByteType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *ByteType:
		return b.Key == o.Key && b.Value == o.Value
	}
	return false
}

// KiloByteType represents a pair of key values in which the value represents bytes
// It converts the value to MB
type KiloByteType struct {
	Key   string
	Value float64
}

// String satisfies the Stringer interface
func (k KiloByteType) String() string {
	return fmt.Sprintf(`"%s":%f`, k.Key, k.Value/KILOBYTE)
}

// Equal compares both keys and values and returns true if they are equal
func (k KiloByteType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *KiloByteType:
		return k.Key == o.Key && k.Value == o.Value
	}
	return false
}

// MegaByteType represents a pair of key values in which the value represents bytes
// It converts the value to MB
type MegaByteType struct {
	Key   string
	Value float64
}

// String satisfies the Stringer interface
func (m MegaByteType) String() string {
	return fmt.Sprintf(`"%s":%f`, m.Key, m.Value/MEGABYTE)
}

// Equal compares both keys and values and returns true if they are equal
func (m MegaByteType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *MegaByteType:
		return m.Key == o.Key && m.Value == o.Value
	}
	return false
}
