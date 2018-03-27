// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"testing"

	recorder_testing "github.com/arsham/expipe/recorder/testing"
)

func TestGetRecorderGoodURL(t *testing.T) {
	url := "http://localhost"
	r := recorder_testing.GetRecorder(url)
	if r == nil {
		t.Error("want (Recorder), got (nil)")
	}
	if r.Name() == "" {
		t.Error("Name cannot be empty")
	}
	if r.IndexName() == "" {
		t.Error("IndexName cannot be empty")
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
	url = "bad url"
	var panicked bool
	func() {
		defer func() {
			if e := recover(); e != nil {
				panicked = true
			}
		}()
		recorder_testing.GetRecorder(url)
		if !panicked {
			t.Error("didn't panic on bad url")
		}
	}()
}
