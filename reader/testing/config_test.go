// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"testing"
	"time"

	"github.com/arsham/expipe/internal"
	rt "github.com/arsham/expipe/reader/testing"
)

func TestConfigNewInstance(t *testing.T) {
	name := "name"
	log := internal.DiscardLogger()
	endpoint := "http://localhost"
	timeout := time.Second
	interval := 100 * time.Millisecond
	backoff := 5
	typeName := "type_name"
	c := rt.Config{
		MockLogger:   log,
		MockName:     name,
		MockEndpoint: endpoint,
		MockTimeout:  timeout,
		MockBackoff:  backoff,
		MockTypeName: typeName,
		MockInterval: interval,
	}
	r, err := c.NewInstance()
	rec, ok := r.(*rt.Reader)
	if !ok {
		t.Error("want (true), got (false)")
	}
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if rec.Name() != c.Name() {
		t.Errorf("want (%v) to be (%v)", c.Name(), rec.Name())
	}
	if rec.Endpoint() != c.Endpoint() {
		t.Errorf("want (%v) to be (%v)", c.Endpoint(), rec.Endpoint())
	}
	if rec.Timeout() != c.Timeout() {
		t.Errorf("want (%v) to be (%v)", c.Timeout(), rec.Timeout())
	}
	if rec.Interval() != c.Interval() {
		t.Errorf("want (%v) to be (%v)", c.Interval(), rec.Interval())
	}
	if rec.Backoff() != c.Backoff() {
		t.Errorf("want (%v) to be (%v)", c.Backoff(), rec.Backoff())
	}
	if rec.TypeName() != c.TypeName() {
		t.Errorf("want (%v) to be (%v)", c.TypeName(), rec.TypeName())
	}
}
