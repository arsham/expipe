// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package lib

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/Sirupsen/logrus"
)

func TestGetLoggerLevels(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		level    string
		expected logrus.Level
	}{
		{"debug", logrus.DebugLevel},
		{"info", logrus.InfoLevel},
		{"warn", logrus.WarnLevel},
		{"error", logrus.ErrorLevel},
		{"DEBUG", logrus.DebugLevel},
		{"INFO", logrus.InfoLevel},
		{"WARN", logrus.WarnLevel},
		{"ERROR", logrus.ErrorLevel},
		{"dEbUG", logrus.DebugLevel},
		{"iNfO", logrus.InfoLevel},
		{"wArN", logrus.WarnLevel},
		{"eRrOR", logrus.ErrorLevel},
		{"", logrus.ErrorLevel},
		{"sdfsdf", logrus.ErrorLevel},
	}

	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			logger := GetLogger(tc.level)
			if logger.Level != tc.expected {
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
