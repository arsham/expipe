// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype_test

import (
	"fmt"
	"reflect"

	"github.com/antonholmquist/jason"
	"github.com/arsham/expvastic/internal/datatype"
)

// This example shows how to read a float64 value into a FloatType
// variable and print it.
func ExampleFromJason_floatType() {
	j, _ := jason.NewValueFromBytes([]byte("666.6"))
	result, err := datatype.FromJason("some float", *j)

	fmt.Printf("error: %v\n", err)
	fmt.Printf("Type: %v\n", reflect.TypeOf(result))

	r := result.(*datatype.FloatType)
	fmt.Printf("Result key: %s\n", r.Key)
	fmt.Printf("Result value: %f\n", r.Value)
	fmt.Printf("String representation: %s\n", string(result.Bytes()))

	// Output:
	// error: <nil>
	// Type: *datatype.FloatType
	// Result key: some float
	// Result value: 666.600000
	// String representation: "some float":666.600000
}

// This example shows how to read a string value into a StringType
// variable and print it.
func ExampleFromJason_stringType() {
	j, _ := jason.NewValueFromBytes([]byte(`"some string"`))
	result, err := datatype.FromJason("string key", *j)
	fmt.Printf("error: %v\n", err)
	fmt.Printf("Type: %v\n", reflect.TypeOf(result))

	r := result.(*datatype.StringType)
	fmt.Printf("Result key: %s\n", r.Key)
	fmt.Printf("Result value: %s\n", r.Value)
	fmt.Printf("String representation: %s\n", string(result.Bytes()))
	// Output:
	// error: <nil>
	// Type: *datatype.StringType
	// Result key: string key
	// Result value: some string
	// String representation: "string key":"some string"
}

// This example shows the error on reading a malformed json object.
func ExampleFromJason_malformedInput() {
	j, _ := jason.NewValueFromBytes([]byte(`{malformed object}`))
	result, err := datatype.FromJason("ignored", *j)

	fmt.Printf("error: %v\n", err)
	fmt.Printf("Type: %v\n", reflect.TypeOf(result))
	// Output:
	// error: unidentified jason value
	// Type: <nil>
}
