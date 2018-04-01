// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package tools

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestGetLoggerLevels(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		level    string
		expected Level
	}{
		{"debug", Level(DebugLevel)},
		{"info", Level(InfoLevel)},
		{"warn", Level(WarnLevel)},
		{"error", Level(ErrorLevel)},
		{"DEBUG", Level(DebugLevel)},
		{"INFO", Level(InfoLevel)},
		{"WARN", Level(WarnLevel)},
		{"ERROR", Level(ErrorLevel)},
		{"dEbUG", Level(DebugLevel)},
		{"iNfO", Level(InfoLevel)},
		{"wArN", Level(WarnLevel)},
		{"eRrOR", Level(ErrorLevel)},
		{"", Level(ErrorLevel)},
		{"sdfsdf", Level(ErrorLevel)},
	}

	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			logger := GetLogger(tc.level)
			if Level(logger.Level) != tc.expected {
				t.Errorf("want (%v), got (%v)", tc.expected, logger.Level)
			}
		})
	}
}

func TestGetDiscardLogger(t *testing.T) {
	logger := DiscardLogger()
	if logger.Out != ioutil.Discard {
		t.Errorf("want (ioutil.Discard), got (%v)", logger.Out)
	}
}
