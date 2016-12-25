// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvar

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/arsham/expvastic/lib"
	"github.com/spf13/viper"
)

func TestLoadExpvarSuccess(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")
	log := lib.DiscardLogger()

	input := bytes.NewBuffer([]byte(`
    readers:
        reader1:
            endpoint: http://localhost
            type_name: example_type
            routepath: /debug/vars
            interval: 2s
            timeout: 3s
            log_level: info
            backoff: 15
    `))

	v.ReadConfig(input)
	c, err := FromViper(v, log, "reader1", "readers.reader1")
	if err != nil {
		t.Fatalf("want no errors, got (%v)", err)
	}
	if c.TypeName() != "example_type" {
		t.Errorf("want (example_type), got (%v)", c.TypeName())
	}
	if c.Endpoint() != "http://localhost" {
		t.Errorf("want (http://localhost), got (%v)", c.Endpoint())
	}
	if c.RoutePath() != "/debug/vars" {
		t.Errorf("want (/debug/vars), got (%v)", c.RoutePath())
	}
	if c.Interval() != time.Duration(2*time.Second) {
		t.Errorf("want (%v), got (%v)", time.Duration(2*time.Second), c.Interval())
	}
	if c.Timeout() != time.Duration(3*time.Second) {
		t.Errorf("want (%v), got (%v)", time.Duration(3*time.Second), c.Timeout())
	}
	if c.Backoff() != 15 {
		t.Errorf("want (15), got (%v)", c.Backoff())
	}
}

func TestLoadExpvarErrors(t *testing.T) {
	t.Parallel()
	v := viper.New()
	log := lib.DiscardLogger()
	v.SetConfigType("yaml")
	tcs := []struct {
		input *bytes.Buffer
	}{
		{ // 0
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
                type_name: example_type
                interval: 2sq
                timeout: 3s
                backoff: 15
    `)),
		},
		{ // 1
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
                type_name: example_type
                interval: 2s
                timeout: 3sw
                backoff: 15
    `)),
		},
		{ // 2
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
                type_name: example_type
                interval: 2s
                timeout: 3s
                backoff: 1
    `)),
		},
		{ // 3
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
                type_name: example_type
                interval: 2s
                timeout: 3s
                backoff: 20w
    `)),
		},
		{ // 4
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
            type_name: example_type
            routepath: /debug/vars
            interval: 2s
            timeout: 3s
            log_level: info
            backoff: 15
    `)),
		},
		{ // 5
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
            type_name: example_type
            endpoint: http:// bad url
            routepath: /debug/vars
            interval: 2s
            timeout: 3s
            log_level: info
            backoff: 15
    `)),
		},
		{ // 5 No types specified
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
            endpoint: http://localhost
            routepath: /debug/vars
            interval: 2s
            timeout: 3s
            log_level: info
            backoff: 15
    `)),
		},
	}
	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			v.ReadConfig(tc.input)
			c, err := FromViper(v, log, "reader1", "readers.reader1")
			if err == nil {
				t.Fatal("want an errors, got nothing")
			}
			if c != nil {
				t.Errorf("want nil conf, got (%v)", c)
			}
		})
	}
}
