// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"testing"
	"time"

	"github.com/arsham/expipe/internal"
	reader_testing "github.com/arsham/expipe/reader/testing"
)

func TestConfig(t *testing.T) {
	name := "name"
	log := internal.DiscardLogger()
	endpoint := "http://localhost"
	timeout := time.Second
	interval := 100 * time.Millisecond
	backoff := 5
	typeName := "type_name"
	c, err := reader_testing.NewConfig(name, typeName, log, endpoint, interval, timeout, backoff)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}

	if c.Name() != name {
		t.Errorf("want (%v) to be (%v)", name, c.Name())
	}
	if c.Logger() != log {
		t.Errorf("want (%v) to be (%v)", log, c.Logger())
	}
	if c.Endpoint() != endpoint {
		t.Errorf("want (%v) to be (%v)", endpoint, c.Endpoint())
	}
	if c.Timeout() != timeout {
		t.Errorf("want (%v) to be (%v)", timeout, c.Timeout())
	}
	if c.Interval() != interval {
		t.Errorf("want (%v) to be (%v)", interval, c.Interval())
	}
	if c.Backoff() != backoff {
		t.Errorf("want (%v) to be (%v)", backoff, c.Backoff())
	}
	if c.TypeName() != typeName {
		t.Errorf("want (%v) to be (%v)", typeName, c.TypeName())
	}
}

func TestConfigNewInstance(t *testing.T) {
	name := "name"
	log := internal.DiscardLogger()
	endpoint := "http://localhost"
	timeout := time.Second
	interval := 100 * time.Millisecond
	backoff := 5
	typeName := "type_name"
	c, err := reader_testing.NewConfig(name, typeName, log, endpoint, interval, timeout, backoff)
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	r, err := c.NewInstance()
	rec, ok := r.(*reader_testing.Reader)
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
