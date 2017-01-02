// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
	"sync"

	"github.com/antonholmquist/jason"
)

// MapConvertMock is the mocked version of MapConvert.
type MapConvertMock struct {
	GCTypes         []string
	MemoryTypes     map[string]MemTypeMock
	ValuesFunc      func(prefix string, values map[string]*jason.Value) []DataType
	mu              sync.Mutex
	DefaultCovertor Mapper
}

// MemTypeMock is the mocked version of memType.
type MemTypeMock struct {
	memType
}

// Values calls the ValuesFunc if exists, otherwise returns nil.
func (m *MapConvertMock) Values(prefix string, values map[string]*jason.Value) []DataType {
	m.mu.Lock()
	m.DefaultCovertor = DefaultMapper()
	m.mu.Unlock()
	return m.DefaultCovertor.Values(prefix, values)
}

// Copy returns a new copy of the Mapper.
func (m *MapConvertMock) Copy() Mapper {
	newMapper := &MapConvertMock{}
	newMapper.GCTypes = m.GCTypes[:]
	newMapper.ValuesFunc = m.ValuesFunc
	newMapper.MemoryTypes = make(map[string]MemTypeMock, len(m.MemoryTypes))
	for k, v := range m.MemoryTypes {
		newMapper.MemoryTypes[k] = v
	}
	return newMapper
}
