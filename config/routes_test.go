// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config

import (
    "bytes"
    "fmt"
    "testing"

    "github.com/arsham/expvastic/lib"
    "github.com/spf13/viper"
)

func equalSlice(a, b []string) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if !isIn(a[i], b) {
            return false
        }
    }
    return true
}

func TestGetRoutesErrors(t *testing.T) {
    v := viper.New()
    v.SetConfigType("yaml")

    tcs := []struct {
        input   *bytes.Buffer
        section string
    }{
        { // 0
            input: bytes.NewBuffer([]byte(`
    routes:
        route1:
            recorders: rec1
    `)),
            section: "readers",
        },
        { // 1
            input: bytes.NewBuffer([]byte(`
    routes:
        route1:
            readers: read1
    `)),
            section: "recorders",
        },
        { // 2
            input: bytes.NewBuffer([]byte(`
    routes:
        route1:
            recorders:
                - rec1
                - rec2
    `)),
            section: "readers",
        },
        { // 3
            input: bytes.NewBuffer([]byte(`
    routes:
        route1:
            readers:
                - read1
                - read2
    `)),
            section: "recorders",
        },
        { // 4
            input: bytes.NewBuffer([]byte(`
    routes:
        route1:
            readers: red1, red2
            recorders:
                - rec1
                - rec2
    `)),
            section: "readers",
        },
        { // 5
            input: bytes.NewBuffer([]byte(`
    routes:
        route1:
            readers:
                - read1
                - read2
            recorders: rec1, rec2
    `)),
            section: "recorders",
        },
    }

    for i, tc := range tcs {
        name := fmt.Sprintf("case_%d", i)
        t.Run(name, func(t *testing.T) {
            v.ReadConfig(tc.input)
            _, err := getRoutes(v)
            if err == nil {
                t.Fatalf("want an error, got nothing (%s)", err)
            }
            if _, ok := err.(interface {
                Routers()
            }); !ok {
                t.Fatalf("expected RoutersErr error, got (%v)", err)
            }
            val := err.(*routersErr)

            if val.Section != tc.section {
                t.Error("want (%s), got (%v)", tc.section, val.Section)
            }
        })
    }
}

func TestGetRoutesValues(t *testing.T) {

    v := viper.New()
    v.SetConfigType("yaml")

    input := bytes.NewBuffer([]byte(`
    routes:
        route1:
            recorders: rec1
            readers: red1
    `))
    var want []string
    v.ReadConfig(input)
    routes, err := getRoutes(v)
    if err != nil {
        t.Fatalf("want no errors, got (%s)", err)
    }
    for name, route := range routes {
        if name != "route1" {
            t.Errorf("want (route1), got (%s)", name)
        }
        want = []string{"rec1"}
        if !equalSlice(want, route.recorders) {
            t.Errorf("want (%v), got (%v)", want, route.recorders)
        }
        want = []string{"red1"}
        if !equalSlice(want, route.readers) {
            t.Errorf("want (%v), got (%v)", want, route.readers)
        }
    }

    input = bytes.NewBuffer([]byte(`
    routes:
        route1:
            recorders:
                - route1_rec1
                - route1_rec2
            readers: [route1_red1, route1_red2]
        route2:
            recorders: [route2_rec1, route2_rec2]
            readers:
                - route2_red1
                - route2_red2
    `))

    v.ReadConfig(input)
    routes, err = getRoutes(v)
    if err != nil {
        t.Fatalf("want no errors, got (%s)", err)
    }
    for name, route := range routes {
        if !isIn(name, []string{"route1", "route2"}) {
            t.Errorf("want (route1 or route2), got (%s)", name)
        }
        want = []string{name + "_rec1", name + "_rec2"}
        if !equalSlice(want, route.recorders) {
            t.Errorf("want (%#v), got (%#v)", want, route.recorders)
        }
        want = []string{name + "_red1", name + "_red2"}
        if !equalSlice(want, route.readers) {
            t.Errorf("want (%v), got (%v)", want, route.readers)
        }
    }
}

func TestCheckRoutesAgainstReadersRecordersErrors(t *testing.T) {
    v := viper.New()
    log := lib.DiscardLogger()
    v.SetConfigType("yaml")

    tcs := []struct {
        input *bytes.Buffer
        err   error
    }{
        {
            input: bytes.NewBuffer([]byte(`
    readers:
        red1:
            type: expvar
    recorders:
        rec1:
            type: elasticsearch
    routes:
        route1:
            recorders: not_exists
            readers: red1
    `)),
            err: newRoutersErr("routers", "not_exists not in recorders", nil),
        },
        {
            input: bytes.NewBuffer([]byte(`
    readers:
        red1:
            type: expvar
    recorders:
        rec1:
            type: elasticsearch
    routes:
        route1:
            recorders: rec1
            readers: not_exists
    `)),
            err: newRoutersErr("routers", "not_exists not in readers", nil),
        },
        {
            input: bytes.NewBuffer([]byte(`
    readers:
        red1:
            type: expvar
        red2:
            type: expvar
    recorders:
        rec1:
            type: elasticsearch
    routes:
        route1:
            readers: red2
            recorders: red1 # wrong one!
    `)),
            err: newRoutersErr("routers", "red1 not in recorders", nil),
        },
    }

    for i, tc := range tcs {
        name := fmt.Sprintf("case_%d", i)
        t.Run(name, func(t *testing.T) {
            v.ReadConfig(tc.input)
            _, err := LoadYAML(log, v)
            if err.Error() != tc.err.Error() {
                t.Fatalf("want (%v), got (%v)", tc.err, err)
            }
        })
    }
}

func TestCheckRoutesAgainstReadersRecordersPasses(t *testing.T) {
    v := viper.New()
    v.SetConfigType("yaml")

    input := bytes.NewBuffer([]byte(`
    readers:
        red1:
            type: expvar
        red2:
            type: expvar
    recorders:
        rec1:
            type: elasticsearch
        rec2:
            type: elasticsearch
    routes:
        route1:
            recorders:
                - rec1
            readers: [red1, red2]
        route2:
            recorders:
                - rec1
                - rec2
            readers: red1
        route3:
            recorders:
                - rec1
                - rec2
            readers:
                - red1
                - red2
    `))

    v.ReadConfig(input)
    readerKeys, _ := getReaders(v)
    recorderKeys, _ := getRecorders(v)
    routes, _ := getRoutes(v)

    if err := checkAgainstReadRecorders(routes, readerKeys, recorderKeys); err != nil {
        t.Fatalf("want no errors, got (%s)", err)
    }
}
