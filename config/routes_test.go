// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/arsham/expipe/internal"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !internal.StringInSlice(a[i], b) {
			return false
		}
	}
	return true
}

func TestGetRoutesErrors(t *testing.T) {
	t.Parallel()
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
			err = errors.Cause(err)
			if err == nil {
				t.Fatalf("getRoutes(v), err = (%s); want (error)", err)
			}
			if _, ok := err.(*RoutersError); !ok {
				t.Fatalf("err.(*RoutersError) = (%v); want RoutersError error", err)
			}
			val := err.(*RoutersError)

			if val.Section != tc.section {
				t.Errorf("val.Section = (%v); want (%s)", val.Section, tc.section)
			}
		})
	}
}

func TestGetRoutesValues(t *testing.T) {
	t.Parallel()

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
			t.Errorf("name = (%s); want (route1)", name)
		}
		want = []string{"rec1"}
		if !equalSlice(want, route.recorders) {
			t.Errorf("equalSlice(want, route.recorders) = (%v); want (%v)", route.recorders, want)
		}
		want = []string{"red1"}
		if !equalSlice(want, route.readers) {
			t.Errorf("equalSlice(want, route.readers) = (%v); want (%v)", route.readers, want)
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
		if !internal.StringInSlice(name, []string{"route1", "route2"}) {
			t.Errorf("internal.StringInSlice(name, ...) = (%s); want (route1 or route2)", name)
		}
		want = []string{name + "_rec1", name + "_rec2"}
		if !equalSlice(want, route.recorders) {
			t.Errorf("equalSlice(want, route.recorders) = (%#v); want (%#v)", route.recorders, want)
		}
		want = []string{name + "_red1", name + "_red2"}
		if !equalSlice(want, route.readers) {
			t.Errorf("equalSlice(want, route.readers) = (%v); want (%v)", route.readers, want)
		}
	}
}

func TestCheckRoutesAgainstReadersRecordersErrors(t *testing.T) {
	t.Parallel()
	v := viper.New()
	log := internal.DiscardLogger()
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
			err: NewRoutersError("routers", "not_exists not in recorders", nil),
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
			err: NewRoutersError("routers", "not_exists not in readers", nil),
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
			err: NewRoutersError("routers", "red1 not in recorders", nil),
		},
	}

	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			v.ReadConfig(tc.input)
			_, err := LoadYAML(log, v)
			if errors.Cause(err).Error() != tc.err.Error() {
				t.Fatalf("err.Error() = (%v); want (%v)", err, tc.err)
			}
		})
	}
}

func TestCheckRoutesAgainstReadersRecordersPasses(t *testing.T) {
	t.Parallel()
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
		t.Fatalf("checkAgainstReadRecorders() = (%s); want (nil)", err)
	}
}

func TestMapMultiReadersToOneRecorder(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")

	input := bytes.NewBuffer([]byte(`
routes:
    route1:
        readers:
            - app_0
            - app_2
            - app_4
        recorders:
            - elastic_1
    route2:
        readers:
            - app_0
            - app_5
        recorders:
            - elastic_2
    route3:
        readers:
            - app_1
            - app_2
        recorders:
            - elastic_1
    `))

	v.ReadConfig(input)
	routes, _ := getRoutes(v)
	routeMap := mapReadersRecorders(routes)

	if len(routeMap) != 2 { // we have two recorders
		t.Errorf("len(routeMap) = (%d); want (2)", len(routeMap))
	}

	wantKeys := []string{"elastic_1", "elastic_2"}
	for key := range routeMap {
		if !internal.StringInSlice(key, wantKeys) {
			t.Fatalf("(%v) not in (%v)", key, wantKeys)
		}
	}

	wantKeys = []string{"app_0", "app_4", "app_1", "app_2"}
	for _, key := range routeMap["elastic_1"] {
		if !internal.StringInSlice(key, wantKeys) {
			t.Errorf("(%v) not in (%v)", key, wantKeys)
		}
	}

	wantKeys = []string{"app_0", "app_5"}
	for _, key := range routeMap["elastic_2"] {
		if !internal.StringInSlice(key, wantKeys) {
			t.Errorf("(%v) not in (%v)", key, wantKeys)
		}
	}
}
