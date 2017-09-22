// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import "github.com/antonholmquist/jason"

// TODO: bypass the operation that won't be converted in any way. They are not supposed to be read
// and converted back.
// TODO: this operation can happen only once. Lazy load the thing.

// JobResultDataTypes generates a list of DataType and puts them inside the DataContainer.
// It returns errors if unmarshaling is unsuccessful or ErrUnidentifiedJason when the container
// ends up empty.
func JobResultDataTypes(b []byte, mapper Mapper) DataContainer {
	obj, err := jason.NewObjectFromBytes(b)
	if err != nil {
		return &Container{Err: err}
	}
	payload := mapper.Values("", obj.Map())

	if len(payload) == 0 {
		expUnidentifiedJSON.Add(1)
		return &Container{Err: ErrUnidentifiedJason}
	}
	return New(payload)
}

// FromJason returns an instance of DataType from a jason value.
func FromJason(key string, value jason.Value) (DataType, error) {
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
