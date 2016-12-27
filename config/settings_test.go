// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/lib"
	"github.com/spf13/viper"
)

func TestLoadSettingsErrors(t *testing.T) {
	t.Parallel()

	v := viper.New()
	log := lib.DiscardLogger()
	nilErr := &StructureErr{"", "", nil}
	v.SetConfigType("yaml")

	input := bytes.NewBuffer([]byte(""))
	v.ReadConfig(input)
	_, err := LoadYAML(log, v)
	if err != EmptyConfigErr {
		t.Errorf("want (%v), got (%v)", EmptyConfigErr, err)
	}

	input = bytes.NewBuffer([]byte(`
    settings:
        log_level:
            - 123
    `))
	v.ReadConfig(input)
	_, err = LoadYAML(log, v)
	if reflect.TypeOf(err) != reflect.TypeOf(nilErr) {
		t.Errorf("want (%v), got (%v)", EmptyConfigErr, err)
	}

	if !strings.Contains(err.Error(), "log_level") {
		t.Errorf("expecting mention of log_level, got (%v)", err)
	}

	input = bytes.NewBuffer([]byte(`
    settings:
        log_level: debug
    `))
	v.ReadConfig(input)
	LoadYAML(log, v)
	if log.Level != logrus.DebugLevel {
		t.Errorf("loglevel wasn't changed, got (%v)", log.Level)
	}
}

func TestLoadSections(t *testing.T) {
	t.Parallel()
	v := viper.New()
	log := lib.DiscardLogger()
	v.SetConfigType("yaml")

	notSpec := func(t *testing.T, err error, section string) {
		if _, ok := err.(interface {
			NotSpecified()
		}); !ok {
			t.Errorf("expected NotSpecified error, got (%v)", err)
		}
		sec := err.(*notSpecifiedErr)
		if sec.Section != section {
			t.Errorf("want (%s) section, got (%v)", section, sec.Section)
		}
		if !strings.Contains(err.Error(), sec.Section) {
			t.Errorf("expected (%s) in error message, got (%v)", sec.Section, err.Error())
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
			_, err := LoadYAML(log, v)
			notSpec(t, err, tc.section)
		})
	}
}

func TestLoadConfiguration(t *testing.T) {
	t.Parallel()
	v := viper.New()
	log := lib.DiscardLogger()
	v.SetConfigType("yaml")

	input := bytes.NewBuffer([]byte(`
    readers:
        reader_1: # populating to get to the passing tests
            interval: 1s
            timeout: 1s
            endpoint: localhost:8200
            backoff: 9
            type_name: erwer
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
	routeMap := map[string]route{"routes": route{
		readers:   []string{"reader_1"},
		recorders: []string{"recorder_1"},
	}}
	_, err := loadConfiguration(v, log, routeMap, readers, recorders)
	if _, ok := err.(interface {
		NotSupported()
	}); !ok {
		t.Errorf("want InvalidEndpoint, got (%v)", err)
	}

	readers = map[string]string{"reader_1": "expvar"}
	recorders = map[string]string{"recorder_1": "not_exists"}
	_, err = loadConfiguration(v, log, routeMap, readers, recorders)
	if _, ok := err.(interface {
		NotSupported()
	}); !ok {
		t.Errorf("want InvalidEndpoint, got (%v)", err)
	}

	readers = map[string]string{"reader_1": "expvar"}
	recorders = map[string]string{"recorder_1": "elasticsearch"}
	_, err = loadConfiguration(v, log, routeMap, readers, recorders)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}

}
