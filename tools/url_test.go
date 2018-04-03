// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package tools_test

import (
	"fmt"
	"testing"

	"github.com/arsham/expipe/tools"
	"github.com/pkg/errors"
)

func TestSanitiseURLErrors(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		input    string
		expected string
	}{
		{"blah", ""},
		{"http localhost", ""},
		{"http:/localhost", ""},
		{"ttp://localhost", ""},
		{"https:/localhost", ""},
		{"https: localhost", ""},
		{"https localhost", ""},
		{"https://loca lhost", ""},
		{"http://loca lhost", ""},
	}

	for i, tc := range tcs {
		name := fmt.Sprintf("error_case_%d", i)
		t.Run(name, func(t *testing.T) {
			res, err := tools.SanitiseURL(tc.input)
			if res != tc.expected {
				t.Errorf("want (%v), got (%v)", tc.expected, res)
			}
			err = errors.Cause(err)
			if _, ok := err.(tools.InvalidURLError); !ok {
				t.Errorf("want (InvalidURLError) type, got (%v)", err)
			}
			if err.Error() != tools.InvalidURLError(tc.input).Error() {
				t.Errorf("want (%v), got (%v)", tools.InvalidURLError(tc.input), err)
			}
		})
	}
}

func TestSanitiseURLPasses(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		input    string
		expected string
	}{
		{"localhost.com", "http://localhost.com"},
		{"www.google.com", "http://www.google.com"},
		{"http://localhost", "http://localhost"},
		{"https://localhost", "https://localhost"},
		{"https://localhost/a", "https://localhost/a"},
		{"http://127.0.0.1", "http://127.0.0.1"},
		{"https://127.0.0.1", "https://127.0.0.1"},
		{"http://127.0.0.1/a", "http://127.0.0.1/a"},
		{"https://127.0.0.1/a", "https://127.0.0.1/a"},
		{"127.0.0.1", "http://127.0.0.1"},
		{"127.0.0.1/aaa", "http://127.0.0.1/aaa"},
		{"plurinucleate.com/seabeach/overable?a=nonefficient&b=velation#erythraeidae", "http://plurinucleate.com/seabeach/overable?a=nonefficient&b=velation#erythraeidae"},
		{"tabebuia.com/proportionment/myelosyphilis?a=sunlessness&b=harlequinesque#hypsochrome", "http://tabebuia.com/proportionment/myelosyphilis?a=sunlessness&b=harlequinesque#hypsochrome"},
		{"stochastical.com/suppurant/sesquiduplicate?a=gilguy&b=sidereal#impuritanism", "http://stochastical.com/suppurant/sesquiduplicate?a=gilguy&b=sidereal#impuritanism"},
		{"colliform.com/gonimic/oxynaphthoic?a=finland&b=amentiform#inexistence", "http://colliform.com/gonimic/oxynaphthoic?a=finland&b=amentiform#inexistence"},
	}

	for i, tc := range tcs {
		name := fmt.Sprintf("pass_case_%d", i)
		t.Run(name, func(t *testing.T) {
			res, err := tools.SanitiseURL(tc.input)
			if res != tc.expected {
				t.Errorf("want (%v), got (%v)", tc.expected, res)
			}
			if err != nil {
				t.Errorf("want nil, got (%v)", err)
			}
		})
	}
}
