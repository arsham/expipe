// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package app_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/arsham/expipe/cmd/expipe/app"
	"github.com/arsham/expipe/config"
	"github.com/arsham/expipe/internal"
	flags "github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

func TestConfigLogLevel(t *testing.T) {
	p := flags.NewParser(&app.Opts, flags.IgnoreUnknown)
	p.Parse()
	log, _, _ := app.Config()
	if log.Level != internal.InfoLevel {
		t.Errorf("want (info), got (%s)", log.Level)
	}

	os.Setenv("LOGLEVEL", "warn")
	p = flags.NewParser(&app.Opts, flags.IgnoreUnknown)
	p.Parse()

	log, conf, err := app.Config()
	if log.Level != internal.WarnLevel {
		t.Errorf("want (warn), got (%s)", log.Level)
	}
	if errors.Cause(err) != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if conf == nil {
		t.Error("want (Config), got (nil)")
	}
}

func TestConfigFileDoesNotExists(t *testing.T) {
	os.Setenv("CONFIG", "thisfiledoesnotexists")
	p := flags.NewParser(&app.Opts, flags.IgnoreUnknown)
	p.Parse()

	_, conf, err := app.Config()
	if errors.Cause(err) == nil {
		t.Error("want (error), got (nil)")
	}
	if conf != nil {
		t.Errorf("want (nil), got (%v)", conf)
	}
}

// returns the file base name and tear down function
func setup(content []byte) (string, func()) {
	cwd, _ := os.Getwd()
	file, err := ioutil.TempFile(cwd, "yaml")
	if err != nil {
		panic(err)
	}
	oldName := file.Name() //required for viper
	newName := file.Name() + ".yml"
	os.Rename(oldName, newName)
	file.Write(content)
	return path.Base(file.Name()), func() {
		os.Remove(newName)
	}
}
func errTestCases() [][]byte {
	return [][]byte{
		[]byte(``),
		[]byte(`readers:
        recorders:`),
		[]byte(`readers: exp
        recorders: es`),
		[]byte(`app:
            type: expvar
        recorders:
            app2:
                type: elasticsearch
        routes:
            readers: app`),
		[]byte(`app:
            type: expvar
        recorders:
            app2:
                type: elasticsearch
        routes:
            readers: app
            recorders: app2`),
		[]byte(`  app: #malformed
            type: expvar
        recorders:
            app2:
                type: elasticsearch
        routes:
            readers: app`),
		[]byte(`readers:
            my_app: # service name
                type: expvar
                endpoint: localhost:1234
                routepath: /debug/vars
                type_name: my_app
                map_file: maps.yml
                interval: 500ms
                timeout: 3s
                backoff: 10
        recorders:
            elastic1: # service name
                type: elasticsearch
                endpoint: http://127.0.0.1:9200
                index_name: expipe
                timeout: 8s
                backoff: 10
        routes:
            route1:
                readers:
                    - my_app1
                recorders:
                    - elastic1
        `),
		[]byte(`readers:
            my_app: # service name
                type: expvar
                endpoint: localhost:1234
                routepath: /debug/vars
                type_name: my_app
                map_file: maps.yml
                interval: 500ms
                timeout: 3s
                backoff: 10
        recorders:
            elastic1: # service name
                type: elasticsearch
                endpoint: http://127.0.0.1:9200
                index_name: expipe
                timeout: 8s
                backoff: 10
        routes:
            route1:
                readers:
                    - my_app
                recorders:
                    - elastic111
        `),
	}
}

func passingInput() []byte {
	return []byte(`readers:
    my_app: # service name
        type: expvar
        endpoint: localhost:1234
        routepath: /debug/vars
        type_name: my_app
        map_file: maps.yml
        interval: 500ms
        timeout: 3s
        backoff: 10
recorders:
    elastic1: # service name
        type: elasticsearch
        endpoint: http://127.0.0.1:9200
        index_name: expipe
        timeout: 8s
        backoff: 10
routes:
    route1:
        readers:
            - my_app
        recorders:
            - elastic1
`)
}

func TestMainAndFromConfigFileErrors(t *testing.T) {
	for i, tc := range errTestCases() {
		name := fmt.Sprintf("fromFlagsCase_%d", i)
		t.Run(name, func(t *testing.T) {
			filename, teardown := setup(tc)
			defer teardown()
			defer func() {
				os.Unsetenv("CONFIG")
			}()

			os.Setenv("CONFIG", filename)
			p := flags.NewParser(&app.Opts, flags.IgnoreUnknown)
			p.Parse()

			_, result, err := app.Config()
			if err == nil {
				t.Error("want error, got nothing")
			}
			if result != nil {
				t.Errorf("want nil, got (%v)", result)
			}
		})
	}
}

func TestMainAndFromConfigFilePasses(t *testing.T) {
	filename, teardown := setup(passingInput())
	defer teardown()
	defer func() {
		os.Unsetenv("CONFIG")
	}()

	os.Setenv("CONFIG", filename)
	p := flags.NewParser(&app.Opts, flags.IgnoreUnknown)
	p.Parse()

	_, result, err := app.Config()
	if err != nil {
		t.Errorf("want nil, got (%v)", err)
	}
	if reflect.TypeOf(result) != reflect.TypeOf(&config.ConfMap{}) {
		t.Errorf("want config.ConfMap, got (%v)", result)
	}
}

func TestMainAndFromFlagsErrors(t *testing.T) {
	os.Unsetenv("CONFIG")
	fakeDuration, _ := time.ParseDuration("dfdfdf")
	tcs := []struct {
		recorder  string
		timeout   time.Duration
		backoff   int
		indexName string
		reader    string
	}{
		{"", 0, 0, "", ""},
		{"localhost:9200", fakeDuration, 0, "", ""},
		{"localhost:9200", time.Second, 0, "", ""},
		{"localhost:9200", time.Second, 20, "", ""},
		{"localhost:9200", time.Second, 20, "222", ""},
		{"localhost:9200", time.Second, 20, "222", "sss"},
		{"localhost:9200", time.Second, 2, "222", "sss"},
		{"localhost:9200", time.Second, 20, "", "sss"},
		{"localhost:9200", time.Second, 2, "222", "localhost6/dev"},
		{"localhost:9200", fakeDuration, 20, "222", "localhost7/dev"},
	}
	for i, tc := range tcs {
		os.Unsetenv("CONFIG")
		os.Setenv("RECORDER", tc.recorder)
		os.Setenv("TIMEOUT", tc.timeout.String())
		os.Setenv("BACKOFF", strconv.Itoa(tc.backoff))
		os.Setenv("INDEXNAME", tc.indexName)
		os.Setenv("READER", tc.reader)
		name := fmt.Sprintf("fromFlagsCase_%d", i)
		t.Run(name, func(t *testing.T) {
			defer func() {
				os.Unsetenv("CONFIG")
				os.Unsetenv("RECORDER")
				os.Unsetenv("TIMEOUT")
				os.Unsetenv("BACKOFF")
				os.Unsetenv("INDEXNAME")
				os.Unsetenv("READER")
			}()

			p := flags.NewParser(&app.Opts, flags.IgnoreUnknown)
			p.Parse()

			_, result, err := app.Config()
			if err == nil {
				t.Error("want error, got nothing")
			}
			if result != nil {
				t.Errorf("want nil, got (%v)", result)
			}
		})
	}
}

func TestMainAndFromFlagsPasses(t *testing.T) {
	os.Unsetenv("CONFIG")
	os.Setenv("READER", "localhost1:222/dev")
	os.Setenv("RECORDER", "localhost2:9200")
	os.Setenv("TIMEOUT", time.Second.String())
	os.Setenv("BACKOFF", strconv.Itoa(20))
	os.Setenv("INDEX", "222")
	os.Setenv("TYPE", "222")
	defer func() {
		os.Unsetenv("CONFIG")
		os.Unsetenv("READER")
		os.Unsetenv("RECORDER")
		os.Unsetenv("TIMEOUT")
		os.Unsetenv("BACKOFF")
		os.Unsetenv("INDEX")
		os.Unsetenv("TYPE")
	}()
	p := flags.NewParser(&app.Opts, flags.IgnoreUnknown)
	p.Parse()

	_, result, err := app.Config()
	if err != nil {
		t.Errorf("want nil, got (%v)", err)
	}
	if reflect.TypeOf(result) != reflect.TypeOf(&config.ConfMap{}) {
		t.Errorf("want config.ConfMap, got (%v)", result)
	}
}
