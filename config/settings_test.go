// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config

import (
    "bytes"
    "fmt"
    "reflect"
    "testing"

    "github.com/Sirupsen/logrus"
    "github.com/arsham/expvastic/lib"
    "github.com/spf13/viper"
)

func isIn(a string, b []string) bool {
    for _, i := range b {
        if a == i {
            return true
        }
    }
    return false
}

func TestLoadSettingsErrors(t *testing.T) {

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
