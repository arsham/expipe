// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package elasticsearch_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/recorder/elasticsearch"
	"github.com/spf13/viper"
)

func TestLoadElasticsearchSuccess(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")
	log := lib.DiscardLogger()

	input := bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
            endpoint: http://127.0.0.1:9200
            index_name: example_index
            timeout: 10s
            backoff: 15
    `))

	v.ReadConfig(input)
	c1, _ := elasticsearch.FromViper(v, log, "recorder1", "recorders.recorder1")
	c2, err := elasticsearch.NewConfig(log, "name", "http://127.0.0.1:9200", 10*time.Second, 15, "example_index")
	for _, c := range []*elasticsearch.Config{c1, c2} {

		if err != nil {
			t.Fatalf("want no errors, got (%v)", err)
		}
		if c.IndexName() != "example_index" {
			t.Errorf("want (example_index), got (%v)", c.IndexName())
		}
		if c.Endpoint() != "http://127.0.0.1:9200" {
			t.Errorf("want (http://127.0.0.1:9200), got (%v)", c.Endpoint())
		}
		if c.Timeout() != time.Duration(10*time.Second) {
			t.Errorf("want (%v), got (%v)", time.Duration(3*time.Second), c.Timeout())
		}
		if c.Backoff() != 15 {
			t.Errorf("want (15), got (%v)", c.Backoff())
		}
	}
}

func TestLoadElasticsearchErrors(t *testing.T) {
	t.Parallel()
	v := viper.New()
	log := lib.DiscardLogger()
	v.SetConfigType("yaml")
	tcs := []struct {
		input *bytes.Buffer
	}{
		{ // 0
			input: bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
                index_name: example_index
                interval: 2sq
                timeout: 1ms
                backoff: 15
    `)),
		},
		{ // 1
			input: bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
                index_name: example_index
                interval: 2s
                timeout: 3sw
                backoff: 15
    `)),
		},
		{ // 2
			input: bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
                index_name: example_index
                interval: 2s
                timeout: 1ms
                backoff: 1
    `)),
		},
		{ // 3
			input: bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
                index_name: example_index
                interval: 2s
                timeout: 1ms
                backoff: 20w
    `)),
		},
		{ // 4
			input: bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
            index_name: example_index
            routepath: /debug/vars
            interval: 2s
            timeout: 1ms
            log_level: info
            backoff: 15
    `)),
		},
		{ // 5
			input: bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
            index_name: example_index
            endpoint: http:// bad url
            routepath: /debug/vars
            interval: 2s
            timeout: 1ms
            log_level: info
            backoff: 15
    `)),
		},
		{ // 5 No types specified
			input: bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
            endpoint: http://127.0.0.1:9200
            routepath: /debug/vars
            interval: 2s
            timeout: 1ms
            log_level: info
            backoff: 15
    `)),
		},
	}
	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			v.ReadConfig(tc.input)
			c, err := elasticsearch.FromViper(v, log, "recorder1", "recorders.recorder1")
			if err == nil {
				t.Fatal("want an errors, got nothing")
			}
			if c != nil {
				t.Errorf("want nil conf, got (%v)", c)
			}
		})
	}
	c, err := elasticsearch.NewConfig(log, "", "http://127.0.0.1:9200", time.Millisecond, 5, "indexName")
	if err == nil {
		t.Error("want error, got nil")
	}
	if c != nil {
		t.Errorf("want nil, got (%v)", c)
	}
}

func TestNewInstance(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigType("yaml")
	log := lib.DiscardLogger()

	input := bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
            endpoint: http://127.0.0.1:9200
            index_name: example_index
            timeout: 3s
            log_level: info
            backoff: 15
    `))

	v.ReadConfig(input)
	c, err := elasticsearch.FromViper(v, log, "recorder1", "recorders.recorder1")
	if err != nil {
		t.Fatalf("want no errors, got (%v)", err)
	}

	r, err := c.NewInstance(context.Background())
	if err != nil {
		t.Errorf("want nil, got (%v)", err)
	}
	if r == nil {
		t.Error("want recorder, got nil")
	}
}
