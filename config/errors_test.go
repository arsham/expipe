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
		(*config.NotSpecifiedError)(nil),
		(*config.RoutersError)(nil),
	}
	for _, tc := range nilTcs {
		if tc.Error() != config.NilStr {
			t.Errorf("tc.Error() = (%s); want (%s)", tc.Error(), config.NilStr)
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
		config.NewNotSpecifiedError(section, reason, fmt.Errorf(body)),
		config.NewRoutersError(section, reason, fmt.Errorf(body)),
	}

	for _, tc := range tcs {
		if !strings.Contains(tc.Error(), section) {
			t.Errorf("tc.Error() = (%s); want (%s) in error", tc.Error(), section)
		}
		if !strings.Contains(tc.Error(), reason) {
			t.Errorf("tc.Error() = (%s); want (%s) in error", tc.Error(), reason)
		}
		if !strings.Contains(tc.Error(), body) {
			t.Errorf("tc.Error() = (%s); want (%s) in error", tc.Error(), body)
		}

	}
	body = "god"
	err2 := config.NotSupportedError(body)
	if !strings.Contains(err2.Error(), body) {
		t.Errorf("err2.Error() = (%s); want (%s) in error", err2.Error(), body)
	}
}
