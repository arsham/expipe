// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
	"strings"
	"sync"

	"github.com/antonholmquist/jason"
	"github.com/arsham/expipe/tools"
	"github.com/spf13/viper"
)

var (
	once       sync.Once
	defaultMap *MapConvert
)

// MapConvert can produce output from GC string list and memory type input.
type MapConvert struct {
	GCTypes     []string
	MemoryTypes map[string]string
}

type treeReader interface {
	IsSet(key string) bool
	GetStringSlice(key string) []string
	GetStringMapString(key string) map[string]string
}

// MapsFromViper reads from the map file and produces functions for conversion
// used in type decoder. It first reads from the default settings defined in the
// maps.yml in the same folder, then overrides with the user specified mappings.
func MapsFromViper(v treeReader) *MapConvert {
	m := &MapConvert{}
	def := DefaultMapper()
	if v.IsSet("gc_types") {
		m.GCTypes = gcTypes(v, def.GCTypes)
	}
	if v.IsSet("memory_bytes") {
		m.MemoryTypes = memoryTypes(v, def.MemoryTypes)
	}
	return m
}

// DefaultMapper returns a MapConvert object that is populated by the default
// mappings. The data is hard coded in the program, but you can provide your
// mapping file in your configuration file.
func DefaultMapper() *MapConvert {
	once.Do(func() {
		v := viper.New()
		v.SetConfigType("yaml")
		v.ReadConfig(defaultMappings())
		defaultMap = &MapConvert{}
		if v.IsSet("gc_types") {
			defaultMap.GCTypes = gcTypes(v, make([]string, 0))
		}
		if v.IsSet("memory_bytes") {
			defaultMap.MemoryTypes = memoryTypes(v, make(map[string]string))
		}
	})
	return defaultMap
}

func (m *MapConvert) getMemoryTypes(prefix, name string, j *jason.Value) (DataType, bool) {
	var (
		data DataType
		ok   bool
	)
	v, err := j.Float64()
	if err != nil {
		dataTypeErrs.Add(1)
		return nil, false
	}
	b := m.MemoryTypes[strings.ToLower(name)]
	if IsByte(b) {
		data, ok = NewByteType(prefix+name, v), true
	} else if IsKiloByte(b) {
		data, ok = NewKiloByteType(prefix+name, v), true
	} else if IsMegaByte(b) {
		data, ok = NewMegaByteType(prefix+name, v), true
	} else {
		data, ok = nil, false
	}
	return data, ok
}

func (m *MapConvert) arrayValue(prefix, name string, a []*jason.Value) DataType {
	if len(a) == 0 {
		return NewFloatListType(prefix+name, []float64{})
	} else if _, err := a[0].Float64(); err == nil {
		if tools.StringInSlice(name, m.GCTypes) {
			return getGCList(prefix+name, a)
		}
		return getFloatListValues(prefix+name, a)
	}
	return nil
}

// Values returns a slice of DataTypes based on the given name/value inputs. It
// flattens the float list values, therefore you will get multiple values per
// input. If the name is found in memory_bytes map, it will return one of those,
// otherwise it will return a FloatType or StringType if can convert. It will
// return nil if the value is not one of above.
func (m *MapConvert) Values(prefix string, values map[string]*jason.Value) []DataType {
	var results []DataType
	input := make(map[string]jason.Value, len(values))
	for k, v := range values {
		input[k] = *v
	}

	for name, value := range input {
		var result DataType
		if _, ok := m.MemoryTypes[strings.ToLower(name)]; ok {
			result, ok = m.getMemoryTypes(prefix, name, &value)
			if !ok {
				continue
			}
			byteTypeCount.Add(1)
		} else if obj, err := value.Object(); err == nil {
			// we are dealing with nested objects
			results = append(results, m.Values(prefix+name+".", obj.Map())...)
			nestedTypeCount.Add(1)
			continue
		} else if s, err := value.String(); err == nil {
			stringTypeCount.Add(1)
			result = NewStringType(prefix+name, s)
		} else if f, err := value.Float64(); err == nil {
			floatTypeCount.Add(1)
			result = NewFloatType(prefix+name, f)
		} else if arr, err := value.Array(); err == nil {
			// we are dealing with an array object
			result = m.arrayValue(prefix, name, arr)
		} else {
			dataTypeErrs.Add(1)
			continue
		}
		dataTypeObjs.Add(1)
		if result != nil { // TEST: write tests (7)
			results = append(results, result)
		}
	}
	return results
}

// Copy returns a new copy of the Mapper.
func (m *MapConvert) Copy() Mapper {
	newMapper := &MapConvert{}
	newMapper.GCTypes = m.GCTypes[:]
	newMapper.MemoryTypes = make(map[string]string, len(m.MemoryTypes))
	for k, v := range m.MemoryTypes {
		newMapper.MemoryTypes[k] = v
	}
	return newMapper
}

func getGCList(name string, arr []*jason.Value) *GCListType {
	res := make([]uint64, len(arr))
	for i, val := range arr {
		if r, err := val.Float64(); err == nil {
			res[i] = uint64(r)
		}
	}
	gCListTypeCount.Add(1)
	return NewGCListType(name, res)
}

func getFloatListValues(name string, arr []*jason.Value) *FloatListType {
	res := make([]float64, len(arr))
	for i, val := range arr {
		if r, err := val.Float64(); err == nil {
			res[i] = r
		}
	}
	floatListTypeCount.Add(1)
	return NewFloatListType(name, res)
}

// IsByte checks the string string to determine if it is a Byte value.
func IsByte(m string) bool { return m == "b" }

// IsKiloByte checks the string string to determine if it is a KiloByte value.
func IsKiloByte(m string) bool { return m == "kb" }

// IsMegaByte checks the string string to determine if it is a MegaByte value.
func IsMegaByte(m string) bool { return m == "mb" }

func gcTypes(v treeReader, gcTypes []string) []string {
	var result []string
	seen := make(map[string]struct{})

	for _, gcType := range v.GetStringSlice("gc_types") {
		seen[gcType] = struct{}{}
		result = append(result, gcType)
	}
	for _, value := range gcTypes {
		if _, ok := seen[value]; !ok {
			result = append(result, value)
		}
	}
	return result
}

func memoryTypes(v treeReader, memTypes map[string]string) map[string]string {
	result := make(map[string]string, len(memTypes))
	for name, memoryType := range v.GetStringMapString("memory_bytes") {
		result[name] = memoryType
	}
	return result
}
