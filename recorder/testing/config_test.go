// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"testing"
	"time"

	rt "github.com/arsham/expipe/recorder/testing"
	"github.com/arsham/expipe/tools"
)

func TestConfigRecorder(t *testing.T) {
	name := "name"
	log := tools.DiscardLogger()
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

	r, err := c.Recorder()
	rec, ok := r.(*rt.Recorder)
	if !ok {
		t.Error("ok = (false); want (true)")
	}
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if rec.Name() != c.Name() {
		t.Errorf("rec.Name() = (%v); want (%v)", rec.Name(), c.Name())
	}
	if rec.Endpoint() != c.Endpoint() {
		t.Errorf("rec.Endpoint() = (%v); want (%v)", rec.Endpoint(), c.Endpoint())
	}
	if rec.Timeout() != c.Timeout() {
		t.Errorf("rec.Timeout() = (%v); want (%v)", rec.Timeout(), c.Timeout())
	}
	if rec.Backoff() != c.Backoff() {
		t.Errorf("rec.Backoff() = (%v); want (%v)", rec.Backoff(), c.Backoff())
	}
	if rec.IndexName() != c.IndexName() {
		t.Errorf("rec.IndexName() = (%v); want (%v)", rec.IndexName(), c.IndexName())
	}
}
