// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
    "io"

    "github.com/antonholmquist/jason"
    "github.com/arsham/expvastic/lib"
)

func jobResultDataTypes(r io.Reader) DataContainer {
    obj, err := jason.NewObjectFromReader(r)
    if err != nil {
        return &Container{Err: err}
    }
    return getJasonValues("", obj.Map())
}

// getJasonValues flattens the map
func getJasonValues(prefix string, values map[string]*jason.Value) *Container {
    result := new(Container)
    for key, value := range values {
        if lib.IsMBType(key) {
            v, err := value.Float64()
            if err != nil {
                continue
            }
            result.Add(ByteType{prefix + key, v})
        } else if obj, err := value.Object(); err == nil {
            // we are dealing with nested objects
            v := getJasonValues(prefix+key+".", obj.Map())
            // TODO: merge them instead
            result.Add(v.List()...)
        } else if arr, err := value.Array(); err == nil {
            // we are dealing with an array object
            result.Add(getFloatListValues(prefix+key, arr)...)
        } else {
            v, err := getJasonValue(prefix+key, *value)
            if err == nil {
                result.Add(v)
            }
        }
    }
    if result.Len() == 0 {
        return &Container{Err: ErrUnidentifiedJason}
    }
    return result
}

func getJasonValue(key string, value jason.Value) (DataType, error) {
    var (
        err error
        s   string
        f   float64
    )
    if s, err = value.String(); err == nil {
        return &StringType{key, s}, nil
    } else if f, err = value.Float64(); err == nil {
        return &FloatType{key, f}, nil
    }
    return nil, ErrUnidentifiedJason
}

func getFloatListValues(key string, values []*jason.Value) []DataType {
    if len(values) == 0 {
        // empty list
        return []DataType{&FloatListType{key, []float64{}}}
    }

    if lib.IsGCType(key) {
        return getGCListValues(key, values)
    }

    if _, err := values[0].Float64(); err == nil {
        res := make([]float64, len(values))
        for i, val := range values {
            if r, err := val.Float64(); err == nil {
                res[i] = r
            }
        }
        return []DataType{&FloatListType{key, res}}
    }
    return nil
}

func getGCListValues(key string, values []*jason.Value) (result []DataType) {
    if _, err := values[0].Float64(); err == nil {
        res := make([]uint64, len(values))
        for i, val := range values {
            if r, err := val.Float64(); err == nil {
                res[i] = uint64(r)
            }
        }
        result = []DataType{&GCListType{key, res}}
    }
    return
}
