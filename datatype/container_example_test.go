// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype_test

import (
	"bytes"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/arsham/expipe/datatype"
)

func ExampleContainer_Len() {
	l := []datatype.DataType{datatype.FloatType{}, datatype.FloatType{}}
	c := datatype.New(l)
	fmt.Println(c.Len())
	// output:
	// 2
}

func ExampleContainer_Add() {
	c := datatype.New(nil)
	firstElm := datatype.FloatType{}

	c.Add(firstElm)
	fmt.Println(c.Len())

	c.Add(datatype.FloatType{}, datatype.FloatType{})
	fmt.Println(c.Len())
	// output:
	// 1
	// 3
}

func ExampleContainer_Generate_new() {
	t := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	c := datatype.Container{}
	w := new(bytes.Buffer)
	n, err := c.Generate(w, t)
	fmt.Println(err, n)
	fmt.Println("Contents:", w.String())
	// output:
	// <nil> 42
	// Contents: {"@timestamp":"2009-11-10T23:00:00+00:00"}
}

func ExampleContainer_Generate_add() {
	t := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	c := datatype.Container{}
	elm := datatype.FloatType{Key: "key", Value: 66.6}
	c.Add(elm)

	w := new(bytes.Buffer)
	_, err := c.Generate(w, t)
	fmt.Println("With new element:", w.String())
	fmt.Println("Error:", err)

	// output:
	// With new element: {"@timestamp":"2009-11-10T23:00:00+00:00","key":66.600000}
	// Error: <nil>
}

func ExampleContainer_Generate_errors() {
	t := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	c := datatype.Container{}
	c.Add(new(badDataType))
	w := new(bytes.Buffer)
	_, err := c.Generate(w, t)
	fmt.Println(errors.Cause(err))
	// output:
	// DataType Error
}
