// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvar

import (
    "bytes"
    "fmt"
    "testing"
    "time"

    "github.com/spf13/viper"
)

func TestLoadExpvar(t *testing.T) {
    v := viper.New()
    v.SetConfigType("yaml")

    input := bytes.NewBuffer([]byte(`
    readers:
        reader1:
            endpoint: http://localhost
            routepath: /debug/vars
            interval: 2s
            timeout: 3s
            log_level: info
            backoff: 15
    `))

    v.ReadConfig(input)
    c, err := FromViper(v, "reader1", "readers.reader1")
    if err != nil {
        t.Fatalf("want no errors, got (%v)", err)
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

func TestLoadExpvar2(t *testing.T) {
    v := viper.New()
    v.SetConfigType("yaml")
    tcs := []struct {
        input *bytes.Buffer
    }{
        { // 0
            input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
                interval: 2sq
                timeout: 3s
                backoff: 15
    `)),
        },
        { // 1
            input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
                interval: 2s
                timeout: 3sw
                backoff: 15
    `)),
        },
        { // 2
            input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
                interval: 2s
                timeout: 3s
                backoff: 1
    `)),
        },
        { // 3
            input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
                interval: 2s
                timeout: 3s
                backoff: 20w
    `)),
        },
    }
    for i, tc := range tcs {
        name := fmt.Sprintf("case_%d", i)
        t.Run(name, func(t *testing.T) {
            v.ReadConfig(tc.input)
            c, err := FromViper(v, "reader1", "readers.reader1")
            if err == nil {
                t.Fatal("want an errors, got nothing")
            }
            if c != nil {
                t.Errorf("want nil conf, got (%v)", c)
            }
        })
    }

}
