// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/arsham/expipe/config"
)

func TestErrorMessages(t *testing.T) {
	nilTcs := []error{
		(*config.StructureErr)(nil),
		(*config.ErrNotSpecified)(nil),
		(*config.ErrRouters)(nil),
	}
	for _, tc := range nilTcs {
		if tc.Error() != config.NilStr {
			t.Errorf("want (%s), got (%s)", config.NilStr, tc.Error())
		}
	}

	section := "this section"
	reason := "the reason"
	body := "whatever body is there"
	tcs := []error{
		&config.StructureErr{
			Section: section,
			Reason:  reason,
			Err:     fmt.Errorf(body),
		},
		config.NewErrNotSpecified(section, reason, fmt.Errorf(body)),
		config.NewErrRouters(section, reason, fmt.Errorf(body)),
	}

	for _, tc := range tcs {
		if !strings.Contains(tc.Error(), section) {
			t.Errorf("want (%s) in error, got (%s)", section, tc.Error())
		}
		if !strings.Contains(tc.Error(), reason) {
			t.Errorf("want (%s) in error, got (%s)", reason, tc.Error())
		}
		if !strings.Contains(tc.Error(), body) {
			t.Errorf("want (%s) in error, got (%s)", body, tc.Error())
		}

	}
	body = "god"
	err2 := config.ErrNotSupported(body)
	if !strings.Contains(err2.Error(), body) {
		t.Errorf("want (%s) in error, got (%s)", body, err2.Error())
	}
}
