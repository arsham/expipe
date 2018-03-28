// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package self_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/reader/self"
	"github.com/spf13/viper"
)

func TestWithLogger(t *testing.T) {
	l := (internal.FieldLogger)(nil)
	c := new(self.Config)
	err := self.WithLogger(l)(c)
	if err == nil {
		t.Error("want (error), got (nil)")
	}
	l = internal.DiscardLogger()
	err = self.WithLogger(l)(c)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if c.Logger() != l {
		t.Errorf("want (%v), got (%v)", l, c.Logger())
	}
}

// tests any with.. that has a string input
func TestWithStrings(t *testing.T) {
	c := new(self.Config)
	tcs := []struct {
		tcName   string
		input    string
		badInput string
		f        func(string) self.Conf
		check    func() string
	}{
		{"name", "name", "", self.WithName, c.Name},
		{"index name", "name", "", self.WithTypeName, c.TypeName},
		{"bad endpoint", "http://localhost", "bad url", self.WithEndpoint, c.Endpoint},
		{"empty endpoint", "http://localhost", "", self.WithEndpoint, c.Endpoint},
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
			c := new(self.Config)
			err := self.WithViper(tc.v, tc.name, tc.key)(c)
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
	c := new(self.Config)
	err := self.WithViper(v, "recorder1", "recorders.recorder1")(c)
	if err != nil {
		t.Fatalf("want (nil), got (%v)", err)
	}
	if c.Backoff() != 15 {
		t.Errorf("want (%d), got (%d)", 15, c.Backoff())
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
	c := new(self.Config)
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
			err := self.WithViper(v, "recorder1", "recorders.recorder1")(c)
			if err == nil {
				t.Error("want (error), got (nil)")
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	log := internal.DiscardLogger()
	c, err := self.NewConfig(
		self.WithLogger(log),
		self.WithName("name"),
		self.WithTypeName("name"),
		self.WithEndpoint("http://localhost"),
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
		args []self.Conf
	}{
		{"error from conf", []self.Conf{self.WithName("")}},
		{"empty name", []self.Conf{
			self.WithEndpoint("http://localhost"),
			self.WithTypeName("indexName")},
		},
		{"empty index name", []self.Conf{
			self.WithEndpoint("http://localhost"),
			self.WithName("name")},
		},
		{"empty endpoint", []self.Conf{
			self.WithName("name"),
			self.WithTypeName("indexName")},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c, err := self.NewConfig(tc.args...)
			if err == nil {
				t.Error("want (error), got (nil)")
			}
			if c != nil {
				t.Errorf("want (nil), got (%v)", c)
			}
		})
	}
}

func TestWithBackoff(t *testing.T) {
	c := new(self.Config)
	err := self.WithBackoff(666)(c)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if c.Backoff() != 666 {
		t.Errorf("want (%v), got (%v)", 666, c.Backoff())
	}
}

func TestWithInterval(t *testing.T) {
	c := new(self.Config)
	err := self.WithInterval(10 * time.Second)(c)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if c.Interval() != 10*time.Second {
		t.Errorf("want (%v), got (%v)", 10*time.Second, c.Interval())
	}
}

func TestNewInstance(t *testing.T) {
	log := internal.DiscardLogger()
	c, err := self.NewConfig(
		self.WithLogger(log),
		self.WithName("name"),
		self.WithTypeName("name"),
		self.WithEndpoint("http://localhost"),
		self.WithInterval(time.Second),
	)
	self.WithBackoff(0)(c)
	if err != nil {
		t.Fatalf("want (nil), got (%v)", err)
	}
	e, err := c.NewInstance()
	if err == nil {
		t.Error("want (error), got (nil)")
	}
	if e.(*self.Reader) != nil {
		t.Errorf("want (nil), got (%v)", e)
	}

	self.WithBackoff(5)(c)
	e, err = c.NewInstance()
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if e.(*self.Reader) == nil {
		t.Error("want (Reader), got (nil)")
	}
}
