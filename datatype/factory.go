// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
    "expvar"

    "github.com/antonholmquist/jason"
    "github.com/arsham/expvastic/lib"
)

var (
    datatypeObjs     = expvar.NewInt("DataType Objects")
    datatypeErrs     = expvar.NewInt("DataType Objects Errors")
    unidentifiedJSON = expvar.NewInt("Unidentified JSON Count")
)

// getJasonValues flattens the map
// The value of ok is true if the object was successully created.
// In each recursion, it prepends the previous key with a period.
// TODO: create a map file for these setup
// Please note that we can't return an error here, it doesn't provide the nested
// elements correctly. Refactoring needed.
func getJasonValues(prefix string, values map[string]*jason.Value) *Container {
    result := new(Container)
    for key, value := range values {
        if lib.IsMBType(key) {
            v, err := value.Float64()
            if err != nil {
                datatypeErrs.Add(1)
                continue
            }
            result.Add(ByteType{prefix + key, v})
            datatypeObjs.Add(1)
        } else if obj, err := value.Object(); err == nil {
            // we are dealing with nested objects
            v := getJasonValues(prefix+key+".", obj.Map())
            result.Add(v.List()...)
        } else if arr, err := value.Array(); err == nil {
            // we are dealing with an array object
            result.Add(floatListValues(prefix+key, arr)...)
            datatypeObjs.Add(1)
        } else {
            v, err := FromJason(prefix+key, *value)
            if err != nil {
                datatypeErrs.Add(1)
                continue
            }
            result.Add(v)
            datatypeObjs.Add(1)
        }
    }
    if result.Len() == 0 {
        unidentifiedJSON.Add(1)
        return &Container{Err: ErrUnidentifiedJason}
    }
    return result
}
