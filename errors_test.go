// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/arsham/expvastic"
)

func TestPingError(t *testing.T) {
	name := "divine"
	err := fmt.Errorf("is a myth")
	e := expvastic.ErrPing{Name: name, Err: err}

	if !strings.Contains(e.Error(), name) {
		t.Errorf("want (%s) in error message, got (%s)", name, e.Error())
	}
	if !strings.Contains(e.Error(), err.Error()) {
		t.Errorf("want (%s) in error message, got (%s)", err.Error(), e.Error())
	}
}
