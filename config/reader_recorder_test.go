// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/arsham/expvastic/lib"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func TestTypeCheckErrors(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")
	log := lib.DiscardLogger()

	input := bytes.NewBuffer([]byte(`
    readers:
        reader1:
            no_type: true
    recorders:
        recorder1:
            type: elasticsearch
    `))
	v.ReadConfig(input)
	_, err := LoadYAML(log, v)

	var (
		val *notSpecifiedErr
		ok  bool
	)
	if val, ok = errors.Cause(err).(*notSpecifiedErr); !ok {
		t.Fatalf("want notSpecifiedErr, got (%v)", err)
	}

	if !strings.Contains(val.Section, "reader1") {
		t.Errorf("want error for (reader1) section, got for (%s)", val.Section)
	}

	input = bytes.NewBuffer([]byte(`
    readers:
        reader1:
            type: expvar
    recorders:
        recorder1:
            no_type: true
    `))
	v.ReadConfig(input)
	_, err = LoadYAML(log, v)

	if val, ok = errors.Cause(err).(*notSpecifiedErr); !ok {
		t.Fatalf("want notSpecifiedErr, got (%v)", err)
	}

	if val.Section != "recorder1" {
		t.Errorf("want error for (recorder1) section, got for (%s)", val.Section)
	}
}

func TestGetReaderKeys(t *testing.T) {
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
		if !lib.StringInSlice(rKey, target) {
			t.Errorf("expected (%s) be in %v", rKey, target)
		}
	}
}

func TestGetKnownReaderKeyTypes(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")

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

func TestGetRecorderKeys(t *testing.T) {
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
		if !lib.StringInSlice(rKey, target) {
			t.Errorf("expected (%s) be in %v", rKey, target)
		}
	}
}

func TestGetKnownRecorderKeyTypes(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")

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
