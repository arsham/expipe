// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder_test

import (
	"fmt"

	. "github.com/arsham/expipe/recorder"
	recorder "github.com/arsham/expipe/recorder/testing"
	"github.com/arsham/expipe/tools"
	"github.com/pkg/errors"
)

func ExampleWithLogger() {
	r := recorder.Recorder{}
	err := WithLogger(nil)(&r)
	fmt.Println("Error:", err == ErrNillLogger)

	err = WithLogger(tools.DiscardLogger())(&r)
	fmt.Println("Error:", err == nil)
	// Output:
	// Error: true
	// Error: true
}

func ExampleWithName() {
	r := recorder.Recorder{}
	err := WithName("")(&r)
	fmt.Println("Error:", err == ErrEmptyName)

	err = WithName("some name")(&r)
	fmt.Println("Error:", err == nil)
	// Output:
	// Error: true
	// Error: true
}

func ExampleWithEndpoint() {

	r := recorder.Recorder{}
	err := WithEndpoint("")(&r)
	err = errors.Cause(err)
	fmt.Println("Error:", err == ErrEmptyEndpoint)

	err = WithEndpoint("somewhere")(&r)
	err = errors.Cause(err)
	fmt.Println("Error:", err)

	err = WithEndpoint("http://localhost")(&r)
	fmt.Println("Error:", err == nil)

	// Output:
	// Error: true
	// Error: invalid endpoint: somewhere
	// Error: true
}
