// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package elasticsearch_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/arsham/expipe/recorder/elasticsearch"
	"github.com/arsham/expipe/tools"
	"github.com/spf13/viper"
)

func TestWithLogger(t *testing.T) {
	l := (tools.FieldLogger)(nil)
	c := new(elasticsearch.Config)
	err := elasticsearch.WithLogger(l)(c)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	l = tools.DiscardLogger()
	err = elasticsearch.WithLogger(l)(c)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if c.Logger() != l {
		t.Errorf("c.Logger() = (%v); want (%v)", c.Logger(), l)
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
				t.Error("err = (nil); want (error)")
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
		t.Fatalf("err = (%v); want (nil)", err)
	}
	if c.Backoff() != 15 {
		t.Errorf("c.Backoff() = (%d); want (%d)", c.Backoff(), 15)
	}
	if c.Timeout() != 10*time.Second {
		t.Errorf("c.Timeout() = (%d); want (%d)", c.Timeout(), 10*time.Second)
	}
	if c.Endpoint() != "http://127.0.0.1:9200" {
		t.Errorf("c.Endpoint() = (%s); want (http://127.0.0.1:9200)", c.Endpoint())
	}
	if c.IndexName() != "example_index" {
		t.Errorf("c.IndexName() = (%s); want (example_index)", c.IndexName())
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
		t.Fatal("err = (nil); want (error)")
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
		t.Fatal("err = (nil); want (error)")
	}
}

func TestNewConfig(t *testing.T) {
	log := tools.DiscardLogger()
	c, err := elasticsearch.NewConfig(
		elasticsearch.WithLogger(log),
	)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if c == nil {
		t.Error("c = (nil); want (Config)")
	}
}

func TestNewConfigErrors(t *testing.T) {
	c, err := elasticsearch.NewConfig(
		elasticsearch.WithLogger(nil),
	)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if c != nil {
		t.Errorf("c = (%v); want (nil)", c)
	}
}

func TestConfigRecorder(t *testing.T) {
	log := tools.DiscardLogger()
	c, err := elasticsearch.NewConfig(
		elasticsearch.WithLogger(log),
	)
	c.ESName = "name"
	c.ESIndexName = "name"
	c.ESEndpoint = "http://localhost"
	c.ESBackoff = 5
	if err != nil {
		t.Fatalf("err = (%v); want (nil)", err)
	}
	e, err := c.Recorder()
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if e.(*elasticsearch.Recorder) != nil {
		t.Errorf("e = (%v); want (nil)", e)
	}

	c.ConfTimeout = time.Second
	e, err = c.Recorder()
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if e.(*elasticsearch.Recorder) == nil {
		t.Error("e = (nil); want (Recorder)")
	}
}
