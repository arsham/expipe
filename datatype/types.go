// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package datatype contains necessary logic to sanitise a JSON object coming
// from a reader. This package is subjected to change.
package datatype

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	// Byte amount is the same as is read.
	Byte = 1.0
	// KiloByte divides the amount to kilobytes to show smaller value.
	KiloByte = 1024 * Byte
	// MegaByte divides the amount to megabytes to show smaller value.
	MegaByte = 1024 * KiloByte
)

// ErrUnidentifiedJason is an error when the value is not identified.
// It happens when the value is not a string or a float64 types,
// or the container ends up empty.
var ErrUnidentifiedJason = errors.New("unidentified jason value")

// readType holds the content of a type.
// Any type that includes readType should populate the content. default index
// value should be zero
type readType struct {
	content string
	index   int // current reading index
}

// Read includes both Key and Value.
func (r *readType) Read(b []byte) (int, error) {
	if r.index >= len(r.content) {
		return 0, io.EOF
	}
	n := copy(b, r.content[r.index:])
	r.index += n
	return n, nil
}

// Reset resets the content to be empty, but it retains the underlying
// storage for use by future writes.
func (r *readType) Reset() { r.index = 0 }

// FloatType represents a pair of key values that the value is a float64.
type FloatType struct {
	readType
	Key   string
	Value float64
}

// NewFloatType returns a new FloadType object.
func NewFloatType(key string, value float64) *FloatType {
	return &FloatType{
		Key:   key,
		Value: value,
		readType: readType{
			content: fmt.Sprintf(`"%s":%f`, key, value),
		},
	}
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
	readType
	Key   string
	Value string
}

// NewStringType returns a new StringType object.
func NewStringType(key, value string) *StringType {
	return &StringType{
		Key:   key,
		Value: value,
		readType: readType{
			content: fmt.Sprintf(`"%s":"%s"`, key, value),
		},
	}
}

// Equal compares both keys and values and returns true if they are equal.
func (s StringType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *StringType:
		return s.Key == o.Key && s.Value == o.Value
	}
	return false
}

// FloatListType represents a pair of key values. The value is a list of floats.
type FloatListType struct {
	readType
	Key   string
	Value []float64
}

// NewFloatListType returns a new FloatListType object.
func NewFloatListType(key string, value []float64) *FloatListType {
	f := &FloatListType{Key: key, Value: value}
	list := make([]string, len(f.Value))
	for i, v := range f.Value {
		list[i] = fmt.Sprintf("%f", v)
	}
	f.content = fmt.Sprintf(`"%s":[%s]`, f.Key, strings.Join(list, ","))

	return f
}

// Equal compares both keys and all values and returns true if they are equal.
// The values are checked in an unordered fashion.
func (f FloatListType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *FloatListType:
		if f.Key != o.Key {
			return false
		}
		for _, v := range o.Value {
			if !FloatInSlice(v, f.Value) {
				return false
			}
		}
		return true
	}
	return false
}

// GCListType represents a pair of key values of GC list info.
type GCListType struct {
	readType
	Key   string
	Value []uint64
}

// NewGCListType returns a new FloatListType object.
func NewGCListType(key string, value []uint64) *GCListType {
	g := &GCListType{Key: key, Value: value}
	// We are filtering, therefore we don't know the size
	var list []string
	for _, v := range g.Value {
		if v > 0 {
			list = append(list, fmt.Sprintf("%d", v/1000))
		}
	}
	g.content = fmt.Sprintf(`"%s":[%s]`, g.Key, strings.Join(list, ","))
	return g
}

// Equal is not implemented. You should iterate and check yourself.
// Equal compares both keys and values and returns true if they are equal.
func (g GCListType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *GCListType:
		if g.Key != o.Key {
			return false
		}
		for _, v := range o.Value {
			if !Uint64InSlice(v, g.Value) {
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
	readType
	Key   string
	Value float64
}

// NewByteType returns a new ByteType object.
func NewByteType(key string, value float64) *ByteType {
	b := &ByteType{Key: key, Value: value}
	b.content = fmt.Sprintf(`"%s":%f`, b.Key, b.Value/MegaByte)
	return b
}

// Equal compares both keys and values and returns true if they are equal.
func (bt ByteType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *ByteType:
		return bt.Key == o.Key && bt.Value == o.Value
	}
	return false
}

// KiloByteType represents a pair of key values in which the value represents
// bytes. It converts the value to KB.
type KiloByteType struct {
	readType
	Key   string
	Value float64
}

// NewKiloByteType returns a new KiloByteType object.
func NewKiloByteType(key string, value float64) *KiloByteType {
	b := &KiloByteType{Key: key, Value: value}
	b.content = fmt.Sprintf(`"%s":%f`, b.Key, b.Value/KiloByte)
	return b
}

// Equal compares both keys and values and returns true if they are equal.
func (k KiloByteType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *KiloByteType:
		return k.Key == o.Key && k.Value == o.Value
	}
	return false
}

// MegaByteType represents a pair of key values in which the value represents
// bytes. It converts the value to MB.
type MegaByteType struct {
	readType
	Key   string
	Value float64
}

// NewMegaByteType returns a new MegaByteType object.
func NewMegaByteType(key string, value float64) *MegaByteType {
	m := &MegaByteType{Key: key, Value: value}
	m.content = fmt.Sprintf(`"%s":%f`, m.Key, m.Value/MegaByte)
	return m
}

// Equal compares both keys and values and returns true if they are equal.
func (m MegaByteType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *MegaByteType:
		return m.Key == o.Key && m.Value == o.Value
	}
	return false
}
