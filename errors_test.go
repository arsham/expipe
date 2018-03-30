// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expipe_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/arsham/expipe"
	"github.com/arsham/expipe/token"
	"github.com/pkg/errors"
)

func TestErrPing(t *testing.T) {
	name := "divine"
	body := "is a myth"
	err := expipe.PingError{name: fmt.Errorf(body)}
	check(t, err.Error(), name)
	check(t, err.Error(), body)

	name2 := "science"
	body2 := "just works!"
	err = expipe.PingError{
		name:  fmt.Errorf(body),
		name2: fmt.Errorf(body2),
	}
	check(t, err.Error(), name)
	check(t, err.Error(), body)
	check(t, err.Error(), name2)
	check(t, err.Error(), body2)
}

func TestErrJob(t *testing.T) {
	name := "divine"
	body := errors.New("is a myth")
	id := token.NewUID()
	err := expipe.JobError{
		ID:   id,
		Name: name,
		Err:  body,
	}
	check(t, err.Error(), name)
	check(t, err.Error(), body.Error())
	check(t, err.Error(), id.String())
}

func check(t *testing.T, err, msg string) {
	if !strings.Contains(err, msg) {
		t.Errorf("Contains(err, msg): want (%s) in (%s)", msg, err)
	}
}
