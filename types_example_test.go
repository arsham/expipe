// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
    "fmt"
    "reflect"

    "github.com/antonholmquist/jason"
    "github.com/arsham/expvastic"
)

func ExampleDataTypeFromJason_floatType() {
    j, _ := jason.NewValueFromBytes([]byte("666.6"))
    result, err := expvastic.DataTypeFromJason("some float", *j)
    fmt.Printf("error: %v\n", err)
    fmt.Printf("Type: %v\n", reflect.TypeOf(result))
    r := result.(*expvastic.FloatType)
    fmt.Printf("Result key: %s\n", r.Key)
    fmt.Printf("Result value: %f\n", r.Value)
    fmt.Printf("String representation: %s\n", result.String())

    // Output:
    // error: <nil>
    // Type: *expvastic.FloatType
    // Result key: some float
    // Result value: 666.600000
    // String representation: "some float":666.600000
}

func ExampleDataTypeFromJason_stringType() {
    j, _ := jason.NewValueFromBytes([]byte(`"some string"`))
    result, err := expvastic.DataTypeFromJason("string key", *j)
    fmt.Printf("error: %v\n", err)
    fmt.Printf("Type: %v\n", reflect.TypeOf(result))
    r := result.(*expvastic.StringType)
    fmt.Printf("Result key: %s\n", r.Key)
    fmt.Printf("Result value: %s\n", r.Value)
    fmt.Printf("String representation: %s\n", result.String())
    // Output:
    // error: <nil>
    // Type: *expvastic.StringType
    // Result key: string key
    // Result value: some string
    // String representation: "string key":"some string"
}

func ExampleDataTypeFromJason_malformedInput() {
    j, _ := jason.NewValueFromBytes([]byte(`{malformed object}`))
    result, err := expvastic.DataTypeFromJason("ignored", *j)
    fmt.Printf("error: %v\n", err)
    fmt.Printf("Type: %v\n", reflect.TypeOf(result))
    // Output:
    // error: unidentified jason value
    // Type: <nil>
}
