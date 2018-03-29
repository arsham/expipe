// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package elasticsearch_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/recorder/elasticsearch"
	"github.com/spf13/viper"
)

func TestWithLogger(t *testing.T) {
	l := (internal.FieldLogger)(nil)
	c := new(elasticsearch.Config)
	err := elasticsearch.WithLogger(l)(c)
	if err == nil {
		t.Error("want (error), got (nil)")
	}
	l = internal.DiscardLogger()
	err = elasticsearch.WithLogger(l)(c)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if c.Logger() != l {
		t.Errorf("want (%v), got (%v)", l, c.Logger())
	}
}

type unmarshaller interface {
	UnmarshalKey(key string, rawVal interface{}) error
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
			c := new(elasticsearch.Config)
			err := elasticsearch.WithViper(tc.v, tc.name, tc.key)(c)
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
            index_name: example_index
            timeout: 10s
            backoff: 15
    `))
	v.ReadConfig(input)
	c := new(elasticsearch.Config)
	err := elasticsearch.WithViper(v, "recorder1", "recorders.recorder1")(c)
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
	if c.IndexName() != "example_index" {
		t.Errorf("want (example_index), got (%s)", c.IndexName())
	}
}

func TestWithViperBadFile(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	input := bytes.NewBuffer([]byte(`
    recorders
        recorder1:
                index_name: example_index
interval: 2sq
                timeout: 1ms
                backoff: 15
    `))
	v.ReadConfig(input)
	c := new(elasticsearch.Config)
	err := elasticsearch.WithViper(v, "recorder1", "recorders.recorder1")(c)
	if err == nil {
		t.Fatal("want (error), got (nil)")
	}

	input = bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
                index_name: example_index
                timeout: asas
                backoff: 15
    `))
	v.ReadConfig(input)
	err = elasticsearch.WithViper(v, "recorder1", "recorders.recorder1")(c)
	if err == nil {
		t.Fatal("want (error), got (nil)")
	}
}

func TestNewConfig(t *testing.T) {
	log := internal.DiscardLogger()
	c, err := elasticsearch.NewConfig(
		elasticsearch.WithLogger(log),
	)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if c == nil {
		t.Error("want (Config), got (nil)")
	}
}

func TestNewInstance(t *testing.T) {
	log := internal.DiscardLogger()
	c, err := elasticsearch.NewConfig(
		elasticsearch.WithLogger(log),
	)
	c.ESName = "name"
	c.ESIndexName = "name"
	c.ESEndpoint = "http://localhost"
	c.ESBackoff = 5
	if err != nil {
		t.Fatalf("want (nil), got (%v)", err)
	}
	e, err := c.NewInstance()
	if err == nil {
		t.Error("want (error), got (nil)")
	}
	if e.(*elasticsearch.Recorder) != nil {
		t.Errorf("want (nil), got (%v)", e)
	}

	c.ConfTimeout = time.Second
	e, err = c.NewInstance()
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if e.(*elasticsearch.Recorder) == nil {
		t.Error("want (Recorder), got (nil)")
	}
}
