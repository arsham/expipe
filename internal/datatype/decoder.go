// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import "github.com/antonholmquist/jason"

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
