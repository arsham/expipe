// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine_test

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"testing/quick"

	"github.com/arsham/expipe/engine"
	"github.com/arsham/expipe/tools/token"
)

func TestPingError(t *testing.T) {
	f := func(names, bodies []string) bool {
		if len(bodies) == 0 {
			bodies = []string{"luFcTbIescdcjouUCWgnfRHgLLqMON3Ty"}
		}
		if len(names) == 0 {
			names = []string{"BtAnEiSCdBbEzgTjCAeoDcMtZsvmuwf4zWIIwuz"}
		}
		max := math.Min(float64(len(names)), float64(len(bodies)))
		for i := 0; i < int(max); i++ {
			err := engine.PingError{names[i]: fmt.Errorf(bodies[i])}
			if !check(t, err.Error(), names[i]) && check(t, err.Error(), bodies[i]) {
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestJobError(t *testing.T) {
	f := func(name, body string) bool {
		id := token.NewUID()
		err := engine.JobError{
			ID:   id,
			Name: name,
			Err:  fmt.Errorf(body),
		}
		return check(t, err.Error(), name) &&
			check(t, err.Error(), body) &&
			check(t, err.Error(), id.String())
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func check(t *testing.T, err, msg string) bool {
	if !strings.Contains(err, msg) {
		t.Errorf("Contains(err, msg): want (%#v) in (%#v)", msg, err)
		return false
	}
	return true
}
