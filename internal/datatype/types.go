// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package datatype contains necessary logic to sanitise a JSON object coming from a reader.
// This package is subjected to change.
package datatype

import (
	"errors"
	"fmt"
	"strings"

	"github.com/arsham/expipe/internal"
)

const (
	// BYTE amount is the same as is read.
	BYTE = 1.0
	// KILOBYTE divides the amount to kilobytes to show smaller value.
	KILOBYTE = 1024 * BYTE
	// MEGABYTE divides the amount to megabytes to show smaller value.
	MEGABYTE = 1024 * KILOBYTE
)

// ErrUnidentifiedJason is an error when the value is not identified.
// It happens when the value is not a string or a float64 types,
// or the container ends up empty.
var ErrUnidentifiedJason = errors.New("unidentified jason value")

// FloatType represents a pair of key values that the value is a float64.
type FloatType struct {
	Key   string
	Value float64
}

// Bytes returns the byte slice representation of the type.
func (f FloatType) Bytes() []byte {
	return []byte(fmt.Sprintf(`"%s":%f`, f.Key, f.Value))
}

// Equal compares both keys and values and returns true if they are equal.
func (f FloatType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *FloatType:
		return f.Key == o.Key && f.Value == o.Value
	}
	return false
}

// StringType represents a pair of key values that the value is a string.
type StringType struct {
	Key   string
	Value string
}

// Bytes returns the byte slice representation of the type.
func (s StringType) Bytes() []byte {
	return []byte(fmt.Sprintf(`"%s":"%s"`, s.Key, s.Value))
}

// Equal compares both keys and values and returns true if they are equal.
func (s StringType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *StringType:
		return s.Key == o.Key && s.Value == o.Value
	}
	return false
}

// FloatListType represents a pair of key values that the value is a list of floats.
type FloatListType struct {
	Key   string
	Value []float64
}

// Bytes returns the byte slice representation of the type.
func (fl FloatListType) Bytes() []byte {
	list := make([]string, len(fl.Value))
	for i, v := range fl.Value {
		list[i] = fmt.Sprintf("%f", v)
	}
	return []byte(fmt.Sprintf(`"%s":[%s]`, fl.Key, strings.Join(list, ",")))
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
			if !internal.FloatInSlice(v, fl.Value) {
				return false
			}
		}
		return true
	}
	return false
}

// GCListType represents a pair of key values of GC list info.
type GCListType struct {
	Key   string
	Value []uint64
}

// Bytes returns the byte slice representation of the type.
func (flt GCListType) Bytes() []byte {
	// We are filtering, therefore we don't know the size
	var list []string
	for _, v := range flt.Value {
		if v > 0 {
			list = append(list, fmt.Sprintf("%d", v/1000))
		}
	}
	return []byte(fmt.Sprintf(`"%s":[%s]`, flt.Key, strings.Join(list, ",")))
}

// Equal is not implemented. You should iterate and check yourself.
// Equal compares both keys and values and returns true if they are equal.
func (flt GCListType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *GCListType:
		if flt.Key != o.Key {
			return false
		}
		for _, v := range o.Value {
			if !internal.Uint64InSlice(v, flt.Value) {
				return false
			}
		}
		return true
	}
	return false
}

// ByteType represents a pair of key values in which the value represents bytes.
// It converts the value to MB.
type ByteType struct {
	Key   string
	Value float64
}

// Bytes returns the byte slice representation of the type.
// It turns the byte into Megabyte.
func (b ByteType) Bytes() []byte {
	return []byte(fmt.Sprintf(`"%s":%f`, b.Key, b.Value/MEGABYTE))
}

// Equal compares both keys and values and returns true if they are equal.
func (b ByteType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *ByteType:
		return b.Key == o.Key && b.Value == o.Value
	}
	return false
}

// KiloByteType represents a pair of key values in which the value represents bytes.
// It converts the value to KB.
type KiloByteType struct {
	Key   string
	Value float64
}

// Bytes returns the byte slice representation of the type.
func (k KiloByteType) Bytes() []byte {
	return []byte(fmt.Sprintf(`"%s":%f`, k.Key, k.Value/KILOBYTE))
}

// Equal compares both keys and values and returns true if they are equal.
func (k KiloByteType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *KiloByteType:
		return k.Key == o.Key && k.Value == o.Value
	}
	return false
}

// MegaByteType represents a pair of key values in which the value represents bytes.
// It converts the value to MB.
type MegaByteType struct {
	Key   string
	Value float64
}

// Bytes returns the byte slice representation of the type.
func (m MegaByteType) Bytes() []byte {
	return []byte(fmt.Sprintf(`"%s":%f`, m.Key, m.Value/MEGABYTE))
}

// Equal compares both keys and values and returns true if they are equal.
func (m MegaByteType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *MegaByteType:
		return m.Key == o.Key && m.Value == o.Value
	}
	return false
}
