// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/arsham/expipe/config"
	"github.com/arsham/expipe/internal"
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

// returns the file base name and tear down function
func setup(content []byte) (string, func()) {
	log = internal.DiscardLogger()
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

func TestMainAndFromConfigFileErrors(t *testing.T) {
	for i, tc := range errTestCases() {
		name := fmt.Sprintf("fromFlagsCase_%d", i)
		t.Run(name, func(t *testing.T) {
			filename, teardown := setup(tc)
			defer teardown()
			result, err := fromConfig(filename)
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
	result, err := fromConfig(filename)
	if err != nil {
		t.Errorf("want nil, got (%v)", err)
	}
	if reflect.TypeOf(result) != reflect.TypeOf(&config.ConfMap{}) {
		t.Errorf("want config.ConfMap, got (%v)", result)
	}
}

func TestMainAndFromFlagsErrors(t *testing.T) {
	opts.ConfFile = ""
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
		opts.Recorder = tc.recorder
		opts.Timeout = tc.timeout
		opts.Backoff = tc.backoff
		opts.IndexName = tc.indexName
		opts.Reader = tc.reader
		name := fmt.Sprintf("fromFlagsCase_%d", i)
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
	opts.ConfFile = ""
	opts.Recorder = "localhost:9200"
	opts.Timeout = time.Second
	opts.Backoff = 20
	opts.IndexName = "222"
	opts.TypeName = "222"
	opts.Reader = "localhost:222/dev"

	result, err := fromFlags()
	if err != nil {
		t.Errorf("want nil, got (%v)", err)
	}
	if reflect.TypeOf(result) != reflect.TypeOf(&config.ConfMap{}) {
		t.Errorf("want config.ConfMap, got (%v)", result)
	}
}
