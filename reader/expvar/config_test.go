// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvar_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader/expvar"
	"github.com/spf13/viper"
)

func TestLoadExpvarSuccess(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")
	log := lib.DiscardLogger()

	input := bytes.NewBuffer([]byte(`
    readers:
        reader1:
            endpoint: http://127.0.0.1:9200
            type_name: example_type
            routepath: /debug/vars
            interval: 2s
            timeout: 3s
            log_level: info
            backoff: 15
    `))

	v.ReadConfig(input)
	c1, _ := expvar.FromViper(v, log, "reader1", "readers.reader1")
	c2, err := expvar.NewConfig(log, "name", "example_type", "http://127.0.0.1:9200", "/debug/vars", 2*time.Second, 3*time.Second, 15, "")
	for _, c := range []*expvar.Config{c1, c2} {

		if err != nil {
			t.Fatalf("want no errors, got (%v)", err)
		}
		if c.TypeName() != "example_type" {
			t.Errorf("want (example_type), got (%v)", c.TypeName())
		}
		if c.Endpoint() != "http://127.0.0.1:9200" {
			t.Errorf("want (http://127.0.0.1:9200), got (%v)", c.Endpoint())
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
}

func TestLoadExpvarErrors(t *testing.T) {
	t.Parallel()
	v := viper.New()
	log := lib.DiscardLogger()
	v.SetConfigType("yaml")
	tcs := []struct {
		input *bytes.Buffer
	}{
		{ // 0
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
                type_name: example_type
                interval: 2sq
                timeout: 3s
                backoff: 15
    `)),
		},
		{ // 1
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
                type_name: example_type
                interval: 2s
                timeout: 3sw
                backoff: 15
    `)),
		},
		{ // 2
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
                type_name: example_type
                interval: 2s
                timeout: 3s
                backoff: 1
    `)),
		},
		{ // 3
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
                type_name: example_type
                interval: 2s
                timeout: 3s
                backoff: 20w
    `)),
		},
		{ // 4
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
            type_name: example_type
            routepath: /debug/vars
            interval: 2s
            timeout: 3s
            log_level: info
            backoff: 15
    `)),
		},
		{ // 5
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
            type_name: example_type
            endpoint: http:// bad url
            routepath: /debug/vars
            interval: 2s
            timeout: 3s
            log_level: info
            backoff: 15
    `)),
		},
		{ // 5 No types specified
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
            endpoint: http://127.0.0.1:9200
            routepath: /debug/vars
            interval: 2s
            timeout: 3s
            log_level: info
            backoff: 15
    `)),
		},
	}
	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			v.ReadConfig(tc.input)
			c, err := expvar.FromViper(v, log, "reader1", "readers.reader1")
			if err == nil {
				t.Fatal("want an errors, got nothing")
			}
			if c != nil {
				t.Errorf("want nil conf, got (%v)", c)
			}
		})
	}

	c, err := expvar.NewConfig(log, "", "example_type", "http://127.0.0.1:9200", "/debug/vars", 2*time.Second, 3*time.Second, 15, "")
	if err == nil {
		t.Error("want error, got nil")
	}
	if c != nil {
		t.Errorf("want nil, got (%v)", c)
	}
}

func TestNewInstance(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	log := lib.DiscardLogger()
	cwd, _ := os.Getwd()
	file, err := ioutil.TempFile(cwd, "yaml")
	if err != nil {
		panic(err)
	}
	file.Write([]byte(`gc_types:
    PauseEnd
    PauseNs
`))
	oldName := file.Name() //required for viper
	newName := file.Name() + ".yml"
	os.Rename(oldName, newName)
	defer os.Remove(newName)

	input := bytes.NewBuffer([]byte(fmt.Sprintf(`
    readers:
        reader1:
            endpoint: http://127.0.0.1:9200
            type_name: example_type
            routepath: /debug/vars
            interval: 2s
            timeout: 3s
            log_level: info
            backoff: 15
            map_file: %s
    `, path.Base(file.Name()))))

	v.ReadConfig(input)
	c, err := expvar.FromViper(v, log, "reader1", "readers.reader1")
	if err != nil {
		t.Fatalf("want no errors, got (%v)", err)
	}

	r, err := c.NewInstance(context.Background())
	if err != nil {
		t.Errorf("want nil, got (%v)", err)
	}
	if r == nil {
		t.Error("want reader, got nil")
	}
	if r.Mapper() == nil {
		t.Error("want mapper, got nil")
	}
}
