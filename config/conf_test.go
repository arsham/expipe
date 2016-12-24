// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config

import (
    "bytes"
    "testing"

    "github.com/arsham/expvastic/lib"
    "github.com/spf13/viper"
)

func TestParseReader(t *testing.T) {
    v := viper.New()
    log := lib.DiscardLogger()
    v.SetConfigType("yaml")

    v.ReadConfig(bytes.NewBuffer([]byte("")))
    _, err := parseReader(v, log, "non_existance_plugin", "readers.reader1")
    if _, ok := err.(NotSupportedErr); !ok {
        t.Errorf("want NotSupportedErr error, got (%v)", err)
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
