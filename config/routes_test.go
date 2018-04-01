// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/arsham/expipe/internal"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func TestGetRoutesErrors(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")
	tcs, err := ReadFixtures("get_routes_errors.txt")
	if err != nil {
		t.Fatalf("error parsing fixture: %v", err)
	}

	for _, tc := range tcs {
		name := fmt.Sprintf("case_%s", tc.Name)
		t.Run(name, func(t *testing.T) {
			v.ReadConfig(tc.Body)
			_, err := getRoutes(v)
			err = errors.Cause(err)
			if err == nil {
				t.Fatalf("getRoutes(v), err = (%s); want (error)", err)
			}
			if _, ok := err.(*RoutersError); !ok {
				t.Fatalf("err.(*RoutersError) = (%v); want RoutersError error", err)
			}
			val := err.(*RoutersError)

			if val.Section != tc.Info {
				t.Errorf("val.Section = (%v); want (%s)", val.Section, tc.Info)
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

	tc, err := FixtureWithSection("various.txt", "GetRoutesValues")
	if err != nil {
		t.Fatalf("error parsing fixture: %v", err)
	}

	v.ReadConfig(tc.Body)
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

	tcs, err := ReadFixtures("check_routes_against_readers_recorders_errors.txt")
	if err != nil {
		t.Fatalf("error parsing fixture: %v", err)
	}

	for _, tc := range tcs {
		name := fmt.Sprintf("case_%s", tc.Name)
		t.Run(name, func(t *testing.T) {
			v.ReadConfig(tc.Body)
			tcErr := NewRoutersError("routers", tc.Info, nil)
			_, err := LoadYAML(log, v)
			_, ok := errors.Cause(err).(*RoutersError)
			if !ok || !strings.Contains(err.Error(), tc.Info) {
				t.Fatalf("err.Error() = (%s); want (%s)", err, tcErr)
			}
		})
	}
}

func TestCheckRoutesAgainstReadersRecordersPasses(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")

	tc, err := FixtureWithSection("various.txt", "CheckRoutesAgainstReadersRecordersPasses")
	if err != nil {
		t.Fatalf("error parsing fixture: %v", err)
	}

	v.ReadConfig(tc.Body)
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

	tc, err := FixtureWithSection("various.txt", "MapMultiReadersToOneRecorder")
	if err != nil {
		t.Fatalf("error parsing fixture: %v", err)
	}

	v.ReadConfig(tc.Body)
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
