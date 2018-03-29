// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"testing"
	"time"

	"github.com/arsham/expipe/internal"
	rt "github.com/arsham/expipe/recorder/testing"
)

func TestConfigNewinstance(t *testing.T) {
	name := "name"
	log := internal.DiscardLogger()
	endpoint := "http://localhost"
	timeout := time.Second
	backoff := 5
	indexName := "index_name"
	c := &rt.Config{
		MockLogger:    log,
		MockName:      name,
		MockEndpoint:  endpoint,
		MockTimeout:   timeout,
		MockBackoff:   backoff,
		MockIndexName: indexName,
	}

	r, err := c.NewInstance()
	rec, ok := r.(*rt.Recorder)
	if !ok {
		t.Error("want (true), got (false)")
	}
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if rec.Name() != c.Name() {
		t.Errorf("want (%v) to be (%v)", rec.Name(), c.Name())
	}
	if rec.Endpoint() != c.Endpoint() {
		t.Errorf("want (%v) to be (%v)", rec.Endpoint(), c.Endpoint())
	}
	if rec.Timeout() != c.Timeout() {
		t.Errorf("want (%v) to be (%v)", rec.Timeout(), c.Timeout())
	}
	if rec.Backoff() != c.Backoff() {
		t.Errorf("want (%v) to be (%v)", rec.Backoff(), c.Backoff())
	}
	if rec.IndexName() != c.IndexName() {
		t.Errorf("want (%v) to be (%v)", rec.IndexName(), c.IndexName())
	}
}
