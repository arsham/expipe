// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config_test

import (
	"fmt"
	"strings"
	"testing"
	"testing/quick"

	"github.com/arsham/expipe/tools/config"
)

func TestErrorMessageNils(t *testing.T) {
	t.Parallel()
	nilTcs := []error{
		(*config.StructureErr)(nil),
		(*config.NotSpecifiedError)(nil),
		(*config.RoutersError)(nil),
	}
	for _, tc := range nilTcs {
		if tc.Error() != config.NilStr {
			t.Errorf("tc.Error() = (%s); want (%s)", tc.Error(), config.NilStr)
		}
	}
}

func TestErrorMessages(t *testing.T) {
	t.Parallel()
	check := func(section, reason, body string) bool {
		tcs := []error{
			&config.StructureErr{Section: section, Reason: reason, Err: fmt.Errorf(body)},
			config.NewNotSpecifiedError(section, reason, fmt.Errorf(body)),
			config.NewRoutersError(section, reason, fmt.Errorf(body)),
		}
		for _, tc := range tcs {
			if !strings.Contains(tc.Error(), section) {
				t.Errorf("section: tc.Error() = (%#v); want (%#v) in error", tc.Error(), section)
				return false
			}
			if !strings.Contains(tc.Error(), reason) {
				t.Errorf("reason: tc.Error() = (%#v); want (%#v) in error", tc.Error(), reason)
				return false
			}
			if !strings.Contains(tc.Error(), body) {
				t.Errorf("body: tc.Error() = (%#v); want (%#v) in error", tc.Error(), body)
				return false
			}
		}
		err := config.NotSupportedError(body)
		if !strings.Contains(err.Error(), body) {
			t.Errorf("body: err.Error() = (%#v); want (%#v) in error", err.Error(), body)
			return false
		}
		return true
	}
	if err := quick.Check(check, nil); err != nil {
		t.Error(err)
	}
}
