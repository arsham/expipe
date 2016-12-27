// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/arsham/expvastic/config"
)

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
                endpoint: 127.0.0.1:9200
                index_name: expvastic
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
                endpoint: 127.0.0.1:9200
                index_name: expvastic
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
        endpoint: 127.0.0.1:9200
        index_name: expvastic
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
	var errMsg string
	cwd, _ := os.Getwd()
	ExitCommand = func(msg string) {
		errMsg = msg
	}
	shallStartEngines = false

	for i, tc := range errTestCases() {
		file, err := ioutil.TempFile(cwd, "yaml")
		if err != nil {
			panic(err)
		}
		oldName := file.Name() //required for viper
		newName := file.Name() + ".yml"
		os.Rename(oldName, newName)
		defer os.Remove(newName)
		file.Write(tc)

		name := fmt.Sprintf("mainCase_%d", i)
		t.Run(name, func(t *testing.T) {
			*confFile = path.Base(file.Name())
			main()
			if errMsg == "" {
				t.Error("want error, got nothing")
			}
		})

		name = fmt.Sprintf("fromFlagsCase_%d", i)
		t.Run(name, func(t *testing.T) {
			result, err := fromConfig(path.Base(file.Name()))
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
	var errMsg string
	cwd, _ := os.Getwd()
	shallStartEngines = false
	ExitCommand = func(msg string) {
		errMsg = msg
	}

	file, err := ioutil.TempFile(cwd, "yaml")
	if err != nil {
		panic(err)
	}
	oldName := file.Name() //required for viper
	newName := file.Name() + ".yml"
	os.Rename(oldName, newName)
	file.Write(passingInput())
	defer os.Remove(newName)

	t.Run("mainCase", func(t *testing.T) {
		*confFile = path.Base(file.Name())
		main()
		if errMsg != "" {
			t.Errorf("want nil, got (%v)", errMsg)
		}
	})

	t.Run("flagCase", func(t *testing.T) {
		result, err := fromConfig(path.Base(file.Name()))
		if err != nil {
			t.Errorf("want nil, got (%v)", err)
		}
		if reflect.TypeOf(result) != reflect.TypeOf(&config.ConfMap{}) {
			t.Errorf("want config.ConfMap, got (%v)", result)
		}
	})
}

func TestMainAndFromFlagsErrors(t *testing.T) {
	var errMsg string
	*confFile = ""
	shallStartEngines = false
	ExitCommand = func(msg string) {
		errMsg = msg
	}
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
		{"localhost:9200", time.Second, 2, "222", "localhost/dev"},
		{"localhost:9200", fakeDuration, 20, "222", "localhost/dev"},
	}
	for i, tc := range tcs {
		*recorder = tc.recorder
		*timeout = tc.timeout
		*backoff = tc.backoff
		*indexName = tc.indexName
		*reader = tc.reader
		name := fmt.Sprintf("mainCase_%d", i)
		t.Run(name, func(t *testing.T) {
			main()
			if errMsg == "" {
				t.Error("want error, got nothing")
			}
		})

		name = fmt.Sprintf("fromFlagsCase_%d", i)
		t.Run(name, func(t *testing.T) {
			result, err := fromFlags()
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
	var errMsg string
	*confFile = ""
	shallStartEngines = false
	ExitCommand = func(msg string) {
		errMsg = msg
	}

	*recorder = "localhost:9200"
	*timeout = time.Second
	*backoff = 20
	*indexName = "222"
	*reader = "localhost:222/dev"

	t.Run("mainCase", func(t *testing.T) {
		main()
		if errMsg != "" {
			t.Errorf("want nil, got (%v)", errMsg)
		}
	})

	t.Run("flagCase", func(t *testing.T) {
		result, err := fromFlags()
		if err != nil {
			t.Errorf("want nil, got (%v)", err)
		}
		if reflect.TypeOf(result) != reflect.TypeOf(&config.ConfMap{}) {
			t.Errorf("want config.ConfMap, got (%v)", result)
		}
	})
}
