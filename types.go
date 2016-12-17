// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
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

// DataType implements Stringer and Marshal/Unmarshal
type DataType interface {
    fmt.Stringer
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

// StringType represents a pair of key values that the value is a string
type StringType struct {
    Key   string
    Value string
}

// String satisfies the Stringer interface
func (s StringType) String() string {
    return fmt.Sprintf(`"%s":"%s"`, s.Key, s.Value)
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

// getDataType constructs a list of DataType values
// First it looks at the key for guessing, then uses the value
// It removes zeros on GC types
func getDataType(key string, value interface{}) (result []DataType) {
    if lib.IsGCType(key) {
        result = unmarshalGCListType(key, value.([]interface{}))
        return
    }
    if lib.IsMBType(key) {
        result = []DataType{&ByteType{key, value.(float64)}}
        return
    }

    switch v := value.(type) {
    case string:
        result = []DataType{&StringType{key, v}}
    case float32, float64:
        result = []DataType{&FloatType{key, v.(float64)}}
    case []interface{}:
        // we have list values
        result = unmarshalFloatListType(key, value.([]interface{}))
    }
    return
}

func unmarshalGCListType(key string, value []interface{}) (result []DataType) {
    if len(value) == 0 {
        // empty list
        result = []DataType{&GCListType{key, []uint64{}}}
        return
    }

    if _, ok := value[0].(float64); ok {
        res := make([]uint64, len(value))
        for i, val := range value {
            res[i] = uint64(val.(float64))
        }
        result = []DataType{&GCListType{key, res}}
    }
    return
}

func unmarshalFloatListType(key string, value []interface{}) (result []DataType) {
    if len(value) == 0 {
        // empty list
        result = []DataType{&FloatListType{key, []float64{}}}
        return
    }

    if _, ok := value[0].(float64); ok {
        res := make([]float64, len(value))
        for i, val := range value {
            res[i] = val.(float64)
        }
        result = []DataType{&FloatListType{key, res}}
    }
    return
}
