// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config_test

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/tools"
	"github.com/arsham/expipe/tools/config"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func TestLoadYAML(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")
	log := tools.DiscardLogger()

	input := bytes.NewBuffer([]byte(`
    readers:
        reader1:
            no_type: true
    recorders:
        recorder1:
            type: elasticsearch
    `))
	v.ReadConfig(input)
	_, err := config.LoadYAML(log, v)

	var (
		val *config.NotSpecifiedError
		ok  bool
	)
	if val, ok = errors.Cause(err).(*config.NotSpecifiedError); !ok {
		t.Fatalf("err.(*config.NotSpecifiedError): err = (%v); want notSpecifiedErr", err)
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
	_, err = config.LoadYAML(log, v)

	if val, ok = errors.Cause(err).(*config.NotSpecifiedError); !ok {
		t.Fatalf("err.(*config.NotSpecifiedError): err = (%v); want (notSpecifiedErr)", err)
	}

	if val.Section != "recorder1" {
		t.Errorf("val.Section = (%s); want (error) for Section", val.Section)
	}
}

func TestLoadYAMLSuccess(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")
	log := tools.DiscardLogger()
	input, err := config.FixtureWithSection("various.txt", "LoadYAMLSuccess")
	if err != nil {
		t.Fatalf("error getting section: %v", err)
	}

	v.ReadConfig(input.Body)
	confMap, err := config.LoadYAML(log, v)
	if errors.Cause(err) != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if confMap == nil {
		t.Error("confMap = (nil); want (confMap)")
	}
}

func stringInMapKeys(niddle string, haystack map[string]reader.DataReader) bool {
	for b := range haystack {
		if b == niddle {
			return true
		}
	}
	return false
}

func TestLoadYAMLRemoveUnused(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")
	log := tools.DiscardLogger()
	input, err := config.FixtureWithSection("various.txt", "LoadYAMLRemoveUnused")
	if err != nil {
		t.Fatalf("error getting section: %v", err)
	}

	v.ReadConfig(input.Body)
	confMap, err := config.LoadYAML(log, v)
	if errors.Cause(err) != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if confMap == nil {
		t.Fatal("confMap = (nil); want (confMap)")
	}
	unwantedReaders := []string{"reader2", "reader3"}
	unwantedRecorders := []string{"recorder3"}
	for red, rec := range confMap.Routes {
		if tools.StringInSlice(red, unwantedReaders) {
			t.Errorf("(%s) reader should not be in (%v))", red, unwantedReaders)
		}
		for _, r := range rec {
			if tools.StringInSlice(r, unwantedRecorders) {
				t.Errorf("(%s) recorder should not be in (%v))", r, unwantedRecorders)
			}
		}
	}
}

func TestLoadSettingsErrors(t *testing.T) {
	t.Parallel()

	v := viper.New()
	log := tools.DiscardLogger()
	nilErr := &config.StructureErr{Section: "", Reason: "", Err: nil}
	v.SetConfigType("yaml")

	input := bytes.NewBuffer([]byte(""))
	v.ReadConfig(input)
	_, err := config.LoadYAML(log, v)
	if err != config.ErrEmptyConfig {
		t.Errorf("want (%v), got (%v)", config.ErrEmptyConfig, err)
	}

	input = bytes.NewBuffer([]byte(`
    settings:
        log_level:
            - 123
    `))
	v.ReadConfig(input)
	_, err = config.LoadYAML(log, v)
	if reflect.TypeOf(err) != reflect.TypeOf(nilErr) {
		t.Errorf("err =  (%v); want (%v)", err, config.ErrEmptyConfig)
	}

	if !strings.Contains(err.Error(), "log_level") {
		t.Errorf("expecting mention of log_level, got (%v)", err)
	}

	input = bytes.NewBuffer([]byte(`
    settings:
        log_level: debug
    `))
	v.ReadConfig(input)
	config.LoadYAML(log, v)
	if log.Level != tools.DebugLevel {
		t.Errorf("log.Level = (%v); want (tools.DebugLevel)", log.Level)
	}
}

func TestLoadSections(t *testing.T) {
	t.Parallel()
	v := viper.New()
	log := tools.DiscardLogger()
	v.SetConfigType("yaml")

	notSpec := func(t *testing.T, err error, section string) {
		if _, ok := errors.Cause(err).(*config.NotSpecifiedError); !ok {
			t.Errorf("err.(*config.NotSpecifiedError) = (%v); want (NotSpecifiedError)", err)
		}

		if !strings.Contains(err.Error(), section) {
			t.Errorf("expected (%s) in error message, got (%v)", section, err.Error())
		}
	}

	tcs := []struct {
		input   *bytes.Buffer
		section string
	}{
		{
			input: bytes.NewBuffer([]byte(`
    readers:
    recorders: blah
    routes: blah
    `)),
			section: "readers",
		},
		{
			input: bytes.NewBuffer([]byte(`
    readers: blah
    recorders:
    routes: blah
    `)),
			section: "recorders",
		},
		{
			input: bytes.NewBuffer([]byte(`
    readers: blah
    recorders: blah
    routes:
    `)),
			section: "routes",
		},
	}
	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			v.ReadConfig(tc.input)
			_, err := config.LoadYAML(log, v)
			notSpec(t, err, tc.section)
		})
	}
}
