// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/arsham/expipe/reader"

	"github.com/arsham/expipe/internal"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func TestLoadConfiguration(t *testing.T) {
	t.Parallel()
	v := viper.New()
	log := internal.DiscardLogger()
	v.SetConfigType("yaml")

	input := bytes.NewBuffer([]byte(`
    readers:
        reader_1: # populating to get to the passing tests
            type_name: expvar
            interval: 1s
            timeout: 1s
            endpoint: localhost:8200
            backoff: 9
        reader_2:
            type_name: self
            interval: 1s
            timeout: 1s
            endpoint: localhost:8200
            backoff: 9
    recorders:
        recorder_1:
            interval: 1s
            timeout: 1s
            endpoint: localhost:8200
            backoff: 9
            index_name: erwer
    routes: blah
    `))
	v.ReadConfig(input)

	readers := map[string]string{"reader_1": "not_exists"}
	recorders := map[string]string{"recorder_1": "elasticsearch"}
	routeMap := map[string]route{"routes": {
		readers:   []string{"reader_1"},
		recorders: []string{"recorder_1"},
	}}
	_, err := loadConfiguration(v, log, routeMap, readers, recorders)
	if _, ok := errors.Cause(err).(ErrNotSupported); !ok {
		t.Errorf("want ErrNotSupported, got (%T)", err)
	}

	readers = map[string]string{"reader_1": "expvar"}
	recorders = map[string]string{"recorder_1": "not_exists"}
	_, err = loadConfiguration(v, log, routeMap, readers, recorders)
	if _, ok := errors.Cause(err).(ErrNotSupported); !ok {
		t.Errorf("want ErrNotSupported, got (%T)", err)
	}

	readers = map[string]string{"reader_1": "expvar", "reader_2": "self"}
	recorders = map[string]string{"recorder_2": "elasticsearch"}
	_, err = loadConfiguration(v, log, routeMap, readers, recorders)
	if err == nil {
		t.Error("want (error), got (nil)")
	}

	readers = map[string]string{"reader_1": "expvar", "reader_2": "self"}
	recorders = map[string]string{"recorder_1": "elasticsearch"}
	_, err = loadConfiguration(v, log, routeMap, readers, recorders)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}

func TestParseReader(t *testing.T) {
	t.Parallel()
	v := viper.New()
	log := internal.DiscardLogger()
	v.SetConfigType("yaml")

	v.ReadConfig(bytes.NewBuffer([]byte("")))
	_, err := parseReader(v, log, "non_existence_plugin", "readers.reader1")
	if _, ok := errors.Cause(err).(ErrNotSupported); !ok {
		t.Errorf("want (ErrNotSupported), got (%v)", err)
	}
	if !strings.Contains(err.Error(), "non_existence_plugin") {
		t.Errorf("expected (non_existence_plugin) in error message, got (%s)", err)
	}

	_, err = parseReader(v, log, "expvar", "readers.reader1")
	if errors.Cause(err) == nil {
		t.Error("want (error), got (nil)")
	}

	_, err = parseReader(v, log, "self", "readers.reader1")
	if errors.Cause(err) == nil {
		t.Error("want (error), got (nil)")
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

	if _, ok := c.(reader.DataReader); !ok {
		t.Errorf("want (reader.DataReader) type, got (%v)", c)
	}
}

func TestGetReaders(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")

	input := bytes.NewBuffer([]byte(`
    readers:
        reader1:
            type: expvar
        reader2:
            type: expvar
    `))
	v.ReadConfig(input)
	keys, err := getReaders(v)
	if err != nil {
		t.Errorf("want nil, got (%v)", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got (%d)", len(keys))
	}

	target := []string{"reader1", "reader2"}
	for rKey := range keys {
		if !internal.StringInSlice(rKey, target) {
			t.Errorf("expected (%s) be in %v", rKey, target)
		}
	}

	// testing known types

	tcs := []struct {
		input *bytes.Buffer
		value string
	}{
		{
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
            type: expvar
    `)),
			value: "expvar",
		},
		{
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
            type: self
    `)),
			value: "self",
		},
	}

	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			v.ReadConfig(tc.input)
			keys, _ := getReaders(v)
			if len(keys) == 0 {
				t.Fatalf("unexpected return value (%v)", keys)
			}
			for _, v := range keys {
				if v != tc.value {
					t.Errorf("want (%s), got (%s)", tc.value, v)
				}
			}
		})
	}
}

func TestGetRecorders(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")

	input := bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
            type: elasticsearch
        recorder2:
            type: elasticsearch
    `))
	v.ReadConfig(input)
	keys, _ := getRecorders(v)
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got (%d)", len(keys))
	}

	target := []string{"recorder1", "recorder2"}
	for rKey := range keys {
		if !internal.StringInSlice(rKey, target) {
			t.Errorf("expected (%s) be in %v", rKey, target)
		}
	}

	// testing known types

	tcs := []struct {
		input *bytes.Buffer
		value string
	}{
		{
			input: bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
            type: elasticsearch
    `)),
			value: "elasticsearch",
		},
	}

	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			v.ReadConfig(tc.input)
			keys, _ := getRecorders(v)
			if len(keys) == 0 {
				t.Fatalf("unexpected return value (%v)", keys)
			}
			for _, v := range keys {
				if v != tc.value {
					t.Errorf("want (%s), got (%s)", tc.value, v)
				}
			}
		})
	}
}
