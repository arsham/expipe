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
		t.Error("r.(*rt.Reader): ok = (false); want (true)")
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
	if rec.Interval() != c.Interval() {
		t.Errorf("rec.Interval() = (%v); want (%v)", rec.Interval(), c.Interval())
	}
	if rec.Backoff() != c.Backoff() {
		t.Errorf("rec.Backoff() = (%v); want (%v)", rec.Backoff(), c.Backoff())
	}
	if rec.TypeName() != c.TypeName() {
		t.Errorf("rec.TypeName() = (%v); want (%v)", rec.TypeName(), c.TypeName())
	}
}
