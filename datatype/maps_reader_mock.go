// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import "github.com/antonholmquist/jason"

// MapConvertMock is the mocked version of MapConvert
type MapConvertMock struct {
	GCTypes         []string
	MemoryTypes     map[string]MemTypeMock
	ValuesFunc      func(prefix string, values map[string]*jason.Value) []DataType
	DefaultCovertor Mapper
}

// MemTypeMock is the mocked version of memType
type MemTypeMock struct {
	memType
}

// Values calls the ValuesFunc if exists, otherwise returns nil
func (m *MapConvertMock) Values(prefix string, values map[string]*jason.Value) []DataType {
	m.DefaultCovertor = DefaultMapper()
	return m.DefaultCovertor.Values(prefix, values)
}
