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

	"github.com/arsham/expipe/config"
	"github.com/arsham/expipe/internal"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func TestLoadYAML(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")
	log := internal.DiscardLogger()

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
	log := internal.DiscardLogger()
	input := bytes.NewBuffer([]byte(`
    readers:
        reader1:
            type: expvar
            endpoint: localhost:1234
            type_name: my_app
            map_file: maps.yml
            interval: 2s
            timeout: 3s
            backoff: 10
    recorders:
        recorder1:
            type: elasticsearch
            endpoint: http://127.0.0.1:9200
            index_name: index
            timeout: 8s
            backoff: 10
    routes:
        route1:
            readers:
                - reader1
            recorders:
                - recorder1
    `))
	v.ReadConfig(input)
	confMap, err := config.LoadYAML(log, v)
	if errors.Cause(err) != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if confMap == nil {
		t.Error("confMap = (nil); want (confMap)")
	}
}

func TestLoadSettingsErrors(t *testing.T) {
	t.Parallel()

	v := viper.New()
	log := internal.DiscardLogger()
	nilErr := &config.StructureErr{Section: "", Reason: "", Err: nil}
	v.SetConfigType("yaml")

	input := bytes.NewBuffer([]byte(""))
	v.ReadConfig(input)
	_, err := config.LoadYAML(log, v)
	if err != config.EmptyConfigErr {
		t.Errorf("want (%v), got (%v)", config.EmptyConfigErr, err)
	}

	input = bytes.NewBuffer([]byte(`
    settings:
        log_level:
            - 123
    `))
	v.ReadConfig(input)
	_, err = config.LoadYAML(log, v)
	if reflect.TypeOf(err) != reflect.TypeOf(nilErr) {
		t.Errorf("err =  (%v); want (%v)", err, config.EmptyConfigErr)
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
	if log.Level != internal.DebugLevel {
		t.Errorf("log.Level = (%v); want (internal.DebugLevel)", log.Level)
	}
}

func TestLoadSections(t *testing.T) {
	t.Parallel()
	v := viper.New()
	log := internal.DiscardLogger()
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
