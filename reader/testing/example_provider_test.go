// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"testing"

	rt "github.com/arsham/expipe/reader/testing"
)

func TestGetRecorderGoodURL(t *testing.T) {
	url := "http://localhost"
	r := rt.GetReader(url)
	if r == nil {
		t.Error("want (Recorder), got (nil)")
	}
	if r.Name() == "" {
		t.Error("Name cannot be empty")
	}
	if r.TypeName() == "" {
		t.Error("TypeName cannot be empty")
	}
	if r.Logger() == nil {
		t.Error("want (Logger), got (nil)")
	}
	if r.Timeout() <= 0 {
		t.Errorf("negative timeout: (%d)", r.Timeout())
	}
	if r.Backoff() < 5 {
		t.Errorf("low backoff: (%d)", r.Backoff())
	}
	if r.Interval() == 0 {
		t.Error("Back off not set")
	}
	url = "bad url"
	var panicked bool
	func() {
		defer func() {
			if e := recover(); e != nil {
				panicked = true
			}
		}()
		rt.GetReader(url)
		if !panicked {
			t.Error("didn't panic on bad url")
		}
	}()
}
