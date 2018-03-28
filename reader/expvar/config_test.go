// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvar_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/reader/expvar"
	"github.com/spf13/viper"
)

func TestWithLogger(t *testing.T) {
	l := (internal.FieldLogger)(nil)
	c := new(expvar.Config)
	err := expvar.WithLogger(l)(c)
	if err == nil {
		t.Error("want (error), got (nil)")
	}
	l = internal.DiscardLogger()
	err = expvar.WithLogger(l)(c)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if c.Logger() != l {
		t.Errorf("want (%v), got (%v)", l, c.Logger())
	}
}

// tests any with.. that has a string input
func TestWithStrings(t *testing.T) {
	c := new(expvar.Config)
	tcs := []struct {
		tcName   string
		input    string
		badInput string
		f        func(string) expvar.Conf
		check    func() string
	}{
		{"name", "name", "", expvar.WithName, c.Name},
		{"index name", "name", "", expvar.WithTypeName, c.TypeName},
		{"bad endpoint", "http://localhost", "bad url", expvar.WithEndpoint, c.Endpoint},
		{"empty endpoint", "http://localhost", "", expvar.WithEndpoint, c.Endpoint},
	}
	for _, tc := range tcs {
		t.Run(tc.tcName, func(t *testing.T) {
			err := tc.f(tc.badInput)(c)
			if err == nil {
				t.Error("want (error), got (nil)")
			}
			err = tc.f(tc.input)(c)
			if err != nil {
				t.Errorf("want (nil), got (%v)", err)
			}
			if tc.check() != tc.input {
				t.Errorf("want (%v), got (%v)", tc.input, tc.check())
			}
		})
	}
}

type unmarshaller interface {
	UnmarshalKey(key string, rawVal interface{}) error
	AllKeys() []string
}

func TestWithViper(t *testing.T) {
	tcs := []struct {
		tcName string
		name   string
		key    string
		v      unmarshaller
	}{
		{"no name", "", "key", viper.New()},
		{"no key", "name", "", viper.New()},
		{"no viper", "name", "key", nil},
	}

	for _, tc := range tcs {
		t.Run(tc.tcName, func(t *testing.T) {
			c := new(expvar.Config)
			err := expvar.WithViper(tc.v, tc.name, tc.key)(c)
			if err == nil {
				t.Error("want (error), got (nil)")
			}
		})
	}
}

func TestWithViperSuccess(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")

	input := bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
            endpoint: http://127.0.0.1:9200
            type_name: example_type
            map_file: noway
            timeout: 10s
            interval: 1s
            backoff: 15
    `))
	v.ReadConfig(input)
	c := new(expvar.Config)
	err := expvar.WithViper(v, "recorder1", "recorders.recorder1")(c)
	if err != nil {
		t.Fatalf("want (nil), got (%v)", err)
	}
	if c.Backoff() != 15 {
		t.Errorf("want (%d), got (%d)", 15, c.Backoff())
	}
	if c.Timeout() != 10*time.Second {
		t.Errorf("want (%d), got (%d)", 10*time.Second, c.Timeout())
	}
	if c.Endpoint() != "http://127.0.0.1:9200" {
		t.Errorf("want (http://127.0.0.1:9200), got (%s)", c.Endpoint())
	}
	if c.TypeName() != "example_type" {
		t.Errorf("want (example_type), got (%s)", c.TypeName())
	}
}

func TestWithViperBadFile(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	c := new(expvar.Config)
	tcs := []struct {
		name  string
		input *bytes.Buffer
	}{
		{
			name: "timeout",
			input: bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
                index_name: example_index
                timeout: abc
                backoff: 15
                interval: 1s
    `)),
		},
		{
			name: "bad interval",
			input: bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
                index_name: example_index
                timeout: 1s
                interval: def
                backoff: 15
    `)),
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			v.ReadConfig(tc.input)
			err := expvar.WithViper(v, "recorder1", "recorders.recorder1")(c)
			if err == nil {
				t.Error("want (error), got (nil)")
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	log := internal.DiscardLogger()
	c, err := expvar.NewConfig(
		expvar.WithLogger(log),
		expvar.WithName("name"),
		expvar.WithTypeName("name"),
		expvar.WithEndpoint("http://localhost"),
	)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if c == nil {
		t.Error("want (Config), got (nil)")
	}
}

func TestNewConfigErrors(t *testing.T) {
	tcs := []struct {
		name string
		args []expvar.Conf
	}{
		{"error from conf", []expvar.Conf{expvar.WithName("")}},
		{"empty name", []expvar.Conf{
			expvar.WithEndpoint("http://localhost"),
			expvar.WithTypeName("indexName")},
		},
		{"empty index name", []expvar.Conf{
			expvar.WithEndpoint("http://localhost"),
			expvar.WithName("name")},
		},
		{"empty endpoint", []expvar.Conf{
			expvar.WithName("name"),
			expvar.WithTypeName("indexName")},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c, err := expvar.NewConfig(tc.args...)
			if err == nil {
				t.Error("want (error), got (nil)")
			}
			if c != nil {
				t.Errorf("want (nil), got (%v)", c)
			}
		})
	}
}
func TestWithTimeout(t *testing.T) {
	c := new(expvar.Config)
	err := expvar.WithTimeout(time.Second)(c)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if c.Timeout() != time.Second {
		t.Errorf("want (%v), got (%v)", time.Second, c.Timeout())
	}
}

func TestWithBackoff(t *testing.T) {
	c := new(expvar.Config)
	err := expvar.WithBackoff(666)(c)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if c.Backoff() != 666 {
		t.Errorf("want (%v), got (%v)", 666, c.Backoff())
	}
}

func TestWithMapFile(t *testing.T) {
	c := new(expvar.Config)
	err := expvar.WithMapFile("")(c)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}

	cwd, _ := os.Getwd()
	file, err := ioutil.TempFile(cwd, "yaml")
	if err != nil {
		panic(err)
	}
	oldName := file.Name() //required for viper
	newName := file.Name() + ".yml"
	os.Rename(oldName, newName)
	defer os.Remove(newName)

	err = expvar.WithMapFile(path.Base(file.Name()))(c)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
}

func TestWithInterval(t *testing.T) {
	c := new(expvar.Config)
	err := expvar.WithInterval(10 * time.Second)(c)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if c.Interval() != 10*time.Second {
		t.Errorf("want (%v), got (%v)", 10*time.Second, c.Interval())
	}
}

func TestNewInstance(t *testing.T) {
	log := internal.DiscardLogger()
	c, err := expvar.NewConfig(
		expvar.WithLogger(log),
		expvar.WithName("name"),
		expvar.WithTypeName("name"),
		expvar.WithEndpoint("http://localhost"),
		expvar.WithInterval(time.Second),
		expvar.WithBackoff(5),
	)
	expvar.WithTimeout(0)(c)
	if err != nil {
		t.Fatalf("want (nil), got (%v)", err)
	}
	e, err := c.NewInstance()
	if err == nil {
		t.Error("want (error), got (nil)")
	}
	if e.(*expvar.Reader) != nil {
		t.Errorf("want (nil), got (%v)", e)
	}

	expvar.WithTimeout(time.Second)(c)
	e, err = c.NewInstance()
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if e.(*expvar.Reader) == nil {
		t.Error("want (Reader), got (nil)")
	}
}
