// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expipe_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/arsham/expipe"
)

func TestPingError(t *testing.T) {
	name1 := "divine"
	err1 := fmt.Errorf("is a myth")
	e := expipe.ErrPing{name1: err1}

	if !strings.Contains(e.Error(), err1.Error()) {
		t.Errorf("want (%s) in error message, got (%s)", err1.Error(), e.Error())
	}

	name2 := "science"
	err2 := fmt.Errorf("is wrong")
	e = expipe.ErrPing{name1: err1, name2: err2}

	if !strings.Contains(e.Error(), err1.Error()) {
		t.Errorf("want (%s) in error message, got (%s)", err1.Error(), e.Error())
	}
	if !strings.Contains(e.Error(), name1) {
		t.Errorf("want (%s) in error message, got (%s)", name1, e.Error())
	}

	if !strings.Contains(e.Error(), err2.Error()) {
		t.Errorf("want (%s) in error message, got (%s)", err2.Error(), e.Error())
	}
	if !strings.Contains(e.Error(), name2) {
		t.Errorf("want (%s) in error message, got (%s)", name2, e.Error())
	}
}
