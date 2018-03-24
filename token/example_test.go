// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package token_test

import (
	"context"
	"fmt"

	"github.com/arsham/expipe/token"
)

// This example shows how to create a new job from a context.
func ExampleNew() {
	job := token.New(context.Background())
	fmt.Println(job)
}
