// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
	"expvar"
	"strings"
	"sync"

	"github.com/antonholmquist/jason"
	"github.com/arsham/expvastic/lib"
	"github.com/spf13/viper"
)

var (
	expStringTypeCount    = expvar.NewInt("StringType Count")
	expFloatTypeCount     = expvar.NewInt("FloatType Count")
	expFloatListTypeCount = expvar.NewInt("FloatListType Count")
	expGCListTypeCount    = expvar.NewInt("GCListType Count")
	expByteTypeCount      = expvar.NewInt("ByteType Count")
	expNestedTypeCount    = expvar.NewInt("Nested Type Count")
	expDataTypeObjs       = expvar.NewInt("DataType Objects")
	expDataTypeErrs       = expvar.NewInt("DataType Objects Errors")
	expUnidentifiedJSON   = expvar.NewInt("Unidentified JSON Count")
	once                  sync.Once
	defaultMap            *MapConvert
)

// Mapper generates DataTypes based on the given name/value inputs.
type Mapper interface {
	// Values closes the channel once all input has been exhausted.
	Values(prefix string, values map[string]*jason.Value) []DataType

	// Copy returns a new copy of the Mapper.
	// You should always copy the mapper if you are using it concurrently.
	Copy() Mapper
}

type memType string

// MapConvert can produce output from GC string list and memory type input.
type MapConvert struct {
	gcTypes     []string
	memoryTypes map[string]memType
}

// MapsFromViper reads from the map file and produces functions for conversion
// used in type decoder. It first reads from the default settings defined in
// the maps.yml in the same folder, then overrides with the user specified mappings.
func MapsFromViper(v *viper.Viper) *MapConvert {
	m := &MapConvert{}
	def := DefaultMapper()
	// TODO: make ignore values. It helps get rid of cmdline
	if v.IsSet("gc_types") {
		m.gcTypes = gcTypes(v, def.gcTypes)
	}

	if v.IsSet("memory_bytes") {
		m.memoryTypes = memoryTypes(v, def.memoryTypes)
	}
	return m
}

// DefaultMapper returns a MapConvert object that is populated by the default mappings.
// The data is hard coded in the program, but you can provide your mapping file
// in your configuration file.
func DefaultMapper() *MapConvert {
	once.Do(func() {
		v := viper.New()
		v.SetConfigType("yaml")
		v.ReadConfig(defaultMappings())
		defaultMap = &MapConvert{}
		if v.IsSet("gc_types") {
			defaultMap.gcTypes = gcTypes(v, make([]string, 0))
		}

		if v.IsSet("memory_bytes") {
			defaultMap.memoryTypes = memoryTypes(v, make(map[string]memType))
		}
	})
	return defaultMap
}

func (m *MapConvert) getMemoryTypes(prefix, name string, value *jason.Value) (DataType, bool) {
	v, err := value.Float64()
	if err != nil {
		expDataTypeErrs.Add(1)
		return nil, false
	}
	b := m.memoryTypes[strings.ToLower(name)]
	if b.IsByte() {
		return &ByteType{prefix + name, v}, true
	} else if b.IsKiloByte() {
		return &KiloByteType{prefix + name, v}, true
	} else if b.IsMegaByte() {
		return &MegaByteType{prefix + name, v}, true
	}
	return nil, false
}

func (m *MapConvert) getArrayValue(prefix, name string, arr []*jason.Value) DataType {
	if len(arr) == 0 {
		return &FloatListType{prefix + name, []float64{}}
	} else if _, err := arr[0].Float64(); err == nil {
		if lib.StringInSlice(name, m.gcTypes) {
			return getGCList(prefix+name, arr)
		}
		return getFloatListValues(prefix+name, arr)
	}
	return nil
}

// Values returns a slice of DataTypes based on the given name/value inputs.
// It flattens the float list values, therefore you will get multiple values per input.
// If the name is found in memory_bytes map, it will return one of those, otherwise it
// will return a FloatType or StringType if can convert.
// It will return nil if the value is not one of above.
func (m *MapConvert) Values(prefix string, values map[string]*jason.Value) []DataType {
	var results []DataType
	input := make(map[string]jason.Value, len(values))
	for k, v := range values {
		input[k] = *v
	}

	for name, value := range input {
		var result DataType
		if stringInMapKeys(name, m.memoryTypes) {
			if stringInMapKeys(name, m.memoryTypes) {
				var ok bool
				result, ok = m.getMemoryTypes(prefix, name, &value)
				if !ok {
					continue
				}
				expByteTypeCount.Add(1)
			}

		} else if obj, err := value.Object(); err == nil {
			// we are dealing with nested objects
			results = append(results, m.Values(prefix+name+".", obj.Map())...)
			continue

		} else if s, err := value.String(); err == nil {
			expStringTypeCount.Add(1)
			result = &StringType{prefix + name, s}
		} else if f, err := value.Float64(); err == nil {
			expFloatTypeCount.Add(1)
			result = &FloatType{prefix + name, f}
		} else if arr, err := value.Array(); err == nil {
			// we are dealing with an array object
			result = m.getArrayValue(prefix, name, arr)
		} else {
			expDataTypeErrs.Add(1)
			continue
		}
		expDataTypeObjs.Add(1)
		if result != nil { // TODO: test
			results = append(results, result)
		}
	}

	return results
}

// Copy returns a new copy of the Mapper.
func (m *MapConvert) Copy() Mapper {
	newMapper := &MapConvert{}
	newMapper.gcTypes = m.gcTypes[:]
	newMapper.memoryTypes = make(map[string]memType, len(m.memoryTypes))
	for k, v := range m.memoryTypes {
		newMapper.memoryTypes[k] = v
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
	expGCListTypeCount.Add(1)
	return &GCListType{name, res}
}

func getFloatListValues(name string, arr []*jason.Value) *FloatListType {
	res := make([]float64, len(arr))
	for i, val := range arr {
		if r, err := val.Float64(); err == nil {
			res[i] = r
		}
	}
	expFloatListTypeCount.Add(1)
	return &FloatListType{name, res}

}

func (m memType) IsByte() bool     { return string(m) == "b" }
func (m memType) IsKiloByte() bool { return string(m) == "kb" }
func (m memType) IsMegaByte() bool { return string(m) == "mb" }

func gcTypes(v *viper.Viper, gcTypes []string) []string {
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

func memoryTypes(v *viper.Viper, memoryTypes map[string]memType) map[string]memType {
	result := make(map[string]memType, len(memoryTypes))
	for name, memoryType := range v.GetStringMapString("memory_bytes") {
		result[name] = memType(memoryType)
	}

	return result
}

func stringInMapKeys(niddle string, haystack map[string]memType) bool {
	niddle = strings.ToLower(niddle)
	for b := range haystack {
		if strings.ToLower(b) == niddle {
			return true
		}
	}
	return false
}
