// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config

import (
	"bytes"
	"strings"
	"testing"

	"github.com/arsham/expipe/internal"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func TestParseReader(t *testing.T) {
	t.Parallel()
	v := viper.New()
	log := internal.DiscardLogger()
	v.SetConfigType("yaml")

	v.ReadConfig(bytes.NewBuffer([]byte("")))
	_, err := parseReader(v, log, "non_existence_plugin", "readers.reader1")
	if _, ok := errors.Cause(err).(ErrNotSupported); !ok {
		t.Errorf("want ErrNotSupported error, got (%v)", err)
	}
	if !strings.Contains(err.Error(), "non_existence_plugin") {
		t.Errorf("expected non_existence_plugin in error message, got (%s)", err)
	}

	input := bytes.NewBuffer([]byte(`
    readers:
        reader1:
            type: expvar
            type_name: expvar_type
            endpoint: http://localhost
            routepath: /debug/vars
            interval: 2s
            timeout: 3s
            log_level: info
            backoff: 15
    `))

	v.ReadConfig(input)
	c, err := parseReader(v, log, "expvar", "reader1")
	if err != nil {
		t.Errorf("want no errors, got (%v)", err)
	}

	if _, ok := c.(Conf); !ok {
		t.Errorf("want Conf type, got (%v)", c)
	}
}

func TestStructureErr(t *testing.T) {
	section := "section name"
	reason := "reason name"
	err := errors.New("my error")
	s := &StructureErr{
		Section: section,
		Reason:  reason,
		Err:     err,
	}

	if !strings.Contains(s.Error(), section) {
		t.Errorf("want (%s) in the message, got (%s)", section, s.Error())
	}
	if !strings.Contains(s.Error(), reason) {
		t.Errorf("want (%s) in the message, got (%s)", reason, s.Error())
	}
	if !strings.Contains(s.Error(), err.Error()) {
		t.Errorf("want (%s) in the message, got (%s)", err.Error(), s.Error())
	}
	if s.Err != err {
		t.Errorf("want (%v), got (%v)", err, s.Err)
	}

	s = (*StructureErr)(nil)
	if s.Error() != nilStr {
		t.Errorf("want (%s), got (%s)", nilStr, s.Error())
	}
}

func TestNotSpecifiedErr(t *testing.T) {
	section := "section name"
	reason := "reason name"
	err := errors.New("my error")
	s := &ErrNotSpecified{
		Section: section,
		Reason:  reason,
		Err:     err,
	}

	if !strings.Contains(s.Error(), section) {
		t.Errorf("want (%s) in the message, got (%s)", section, s.Error())
	}
	if !strings.Contains(s.Error(), reason) {
		t.Errorf("want (%s) in the message, got (%s)", reason, s.Error())
	}
	if !strings.Contains(s.Error(), err.Error()) {
		t.Errorf("want (%s) in the message, got (%s)", err.Error(), s.Error())
	}
	if s.Err != err {
		t.Errorf("want (%v), got (%v)", err, s.Err)
	}

	s = (*ErrNotSpecified)(nil)
	if s.Error() != nilStr {
		t.Errorf("want (%s), got (%s)", nilStr, s.Error())
	}
}

func TestRoutersErr(t *testing.T) {
	section := "section name"
	reason := "reason name"
	err := errors.New("my error")
	s := NewErrRouters(section, reason, err)

	if !strings.Contains(s.Error(), section) {
		t.Errorf("want (%s) in the message, got (%s)", section, s.Error())
	}
	if !strings.Contains(s.Error(), reason) {
		t.Errorf("want (%s) in the message, got (%s)", reason, s.Error())
	}
	if !strings.Contains(s.Error(), err.Error()) {
		t.Errorf("want (%s) in the message, got (%s)", err.Error(), s.Error())
	}
	if s.Err != err {
		t.Errorf("want (%v), got (%v)", err, s.Err)
	}

	s = (*ErrRouters)(nil)
	if s.Error() != nilStr {
		t.Errorf("want (%s), got (%s)", nilStr, s.Error())
	}
}

func TestNotSupportedErr(t *testing.T) {
	msg := "god"
	s := ErrNotSupported(msg)

	if !strings.Contains(s.Error(), msg) {
		t.Errorf("want (%s) in the message, got (%s)", msg, s.Error())
	}

	if _, ok := interface{}(s).(ErrNotSupported); !ok {
		t.Errorf("want ErrNotSupported interface, got (%v)", s)
	}
}
