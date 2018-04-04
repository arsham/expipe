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

// FloatType represents a pair of key values that the value is a float64.
type FloatType struct {
	Key     string
	Value   float64
	content string
	index   int // current reading index
}

// NewFloatType returns a new FloadType object.
func NewFloatType(key string, value float64) *FloatType {
	return &FloatType{
		Key:     key,
		Value:   value,
		content: fmt.Sprintf(`"%s":%f`, key, value),
		index:   0,
	}
}

// Read includes both Key and Value.
func (f *FloatType) Read(b []byte) (int, error) {
	if f.index >= len(f.content) {
		return 0, io.EOF
	}
	n := copy(b, f.content[f.index:])
	f.index += n
	return n, nil
}

// Reset resets the content to be empty, but it retains the underlying
// storage for use by future writes.
func (f *FloatType) Reset() { f.index = 0 }

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
	Key     string
	Value   string
	content string
	index   int
}

// NewStringType returns a new StringType object.
func NewStringType(key, value string) *StringType {
	return &StringType{
		Key:     key,
		Value:   value,
		content: fmt.Sprintf(`"%s":"%s"`, key, value),
		index:   0,
	}
}

// Read includes both Key and Value.
func (s *StringType) Read(b []byte) (int, error) {
	if s.index >= len(s.content) {
		return 0, io.EOF
	}
	n := copy(b, s.content[s.index:])
	s.index += n
	return n, nil
}

// Reset resets the content to be empty, but it retains the underlying
// storage for use by future writes.
func (s *StringType) Reset() { s.index = 0 }

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
	Key     string
	Value   []float64
	content string
	index   int
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

// Read includes both Key and Value.
func (f *FloatListType) Read(b []byte) (int, error) {
	if f.index >= len(f.content) {
		return 0, io.EOF
	}
	n := copy(b, f.content[f.index:])
	f.index += n
	return n, nil
}

// Reset resets the content to be empty, but it retains the underlying
// storage for use by future writes.
func (f *FloatListType) Reset() { f.index = 0 }

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
	Key     string
	Value   []uint64
	content string
	index   int
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

// Read includes both Key and Value.
func (g *GCListType) Read(b []byte) (int, error) {
	if g.index >= len(g.content) {
		return 0, io.EOF
	}
	n := copy(b, g.content[g.index:])
	g.index += n
	return n, nil
}

// Reset resets the content to be empty, but it retains the underlying
// storage for use by future writes.
func (g *GCListType) Reset() { g.index = 0 }

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
	Key     string
	Value   float64
	content string
	index   int
}

// NewByteType returns a new ByteType object.
func NewByteType(key string, value float64) *ByteType {
	b := &ByteType{Key: key, Value: value}
	b.content = fmt.Sprintf(`"%s":%f`, b.Key, b.Value/MegaByte)
	return b
}

// Read includes both Key and Value.
func (bt *ByteType) Read(b []byte) (int, error) {
	if bt.index >= len(bt.content) {
		return 0, io.EOF
	}
	n := copy(b, bt.content[bt.index:])
	bt.index += n
	return n, nil
}

// Reset resets the content to be empty, but it retains the underlying
// storage for use by future writes.
func (bt *ByteType) Reset() { bt.index = 0 }

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
	Key     string
	Value   float64
	content string
	index   int
}

// NewKiloByteType returns a new KiloByteType object.
func NewKiloByteType(key string, value float64) *KiloByteType {
	b := &KiloByteType{Key: key, Value: value}
	b.content = fmt.Sprintf(`"%s":%f`, b.Key, b.Value/KiloByte)
	return b
}

// Read includes both Key and Value.
func (k *KiloByteType) Read(b []byte) (int, error) {
	if k.index >= len(k.content) {
		return 0, io.EOF
	}
	n := copy(b, k.content[k.index:])
	k.index += n
	return n, nil
}

// Reset resets the content to be empty, but it retains the underlying
// storage for use by future writes.
func (k *KiloByteType) Reset() { k.index = 0 }

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
	Key     string
	Value   float64
	content string
	index   int
}

// NewMegaByteType returns a new MegaByteType object.
func NewMegaByteType(key string, value float64) *MegaByteType {
	m := &MegaByteType{Key: key, Value: value}
	m.content = fmt.Sprintf(`"%s":%f`, m.Key, m.Value/MegaByte)
	return m
}

// Read includes both Key and Value.
func (m *MegaByteType) Read(b []byte) (int, error) {
	if m.index >= len(m.content) {
		return 0, io.EOF
	}
	n := copy(b, m.content[m.index:])
	m.index += n
	return n, nil
}

// Reset resets the content to be empty, but it retains the underlying
// storage for use by future writes.
func (m *MegaByteType) Reset() { m.index = 0 }

// Equal compares both keys and values and returns true if they are equal.
func (m MegaByteType) Equal(other DataType) bool {
	switch o := other.(type) {
	case *MegaByteType:
		return m.Key == o.Key && m.Value == o.Value
	}
	return false
}
