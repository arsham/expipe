// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package self_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/reader/self"
	"github.com/spf13/viper"
)

func TestLoadExpvarSuccess(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")
	log := internal.DiscardLogger()

	input := bytes.NewBuffer([]byte(`
    readers:
        reader1:
            endpoint: http://localhost
            type_name: example_type
            interval: 2s
            backoff: 15
    `))

	v.ReadConfig(input)
	c, err := self.FromViper(v, log, "reader1", "readers.reader1")
	if err != nil {
		t.Fatalf("want no errors, got (%v)", err)
	}
	if c.TypeName() != "example_type" {
		t.Errorf("want (example_type), got (%v)", c.TypeName())
	}
	if c.Interval() != time.Duration(2*time.Second) {
		t.Errorf("want (%v), got (%v)", time.Duration(2*time.Second), c.Interval())
	}
	if c.Backoff() != 15 {
		t.Errorf("want (15), got (%v)", c.Backoff())
	}
}

func TestLoadExpvarErrors(t *testing.T) {
	t.Parallel()
	v := viper.New()
	log := internal.DiscardLogger()
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
                backoff: 15
    `)),
		},
		{ // 1
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
                type_name: example_type
                interval: 2s
                backoff: 20w
    `)),
		},
		{ // 2 No types specified
			input: bytes.NewBuffer([]byte(`
    readers:
        reader1:
            interval: 2s
            timeout: 3s
            backoff: 15
    `)),
		},
	}
	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			v.ReadConfig(tc.input)
			c, err := self.FromViper(v, log, "reader1", "readers.reader1")
			if err == nil {
				t.Fatal("want an errors, got nothing")
			}
			if c != nil {
				t.Errorf("want nil conf, got (%v)", c)
			}
		})
	}
}

func TestNewInstance(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")
	log := internal.DiscardLogger()
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
	ts := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	input := bytes.NewBuffer([]byte(fmt.Sprintf(`
    readers:
        reader1:
            type_name: self
            interval: 2s
            timeout: 2s
            backoff: 15
            map_file: %s
    `, path.Base(file.Name()))))

	v.ReadConfig(input)
	c, err := self.FromViper(v, log, "reader1", "readers.reader1")
	if err != nil {
		t.Fatalf("want no errors, got (%v)", err)
	}
	c.SelfEndpoint = ts.URL
	r, err := c.NewInstance()
	if err != nil {
		t.Errorf("want nil, got (%v)", err)
	}
	if r == nil {
		t.Fatal("want reader, got nil")
	}
	err = r.Ping()
	if err != nil {
		t.Fatal(err)
	}
	if r.Mapper() == nil {
		t.Error("want mapper, got nil")
	}
}
