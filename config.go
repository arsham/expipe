// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
    "time"

    "github.com/Sirupsen/logrus"
    "github.com/arsham/expvastic/reader"
    "github.com/arsham/expvastic/recorder"
)

// Conf holds the necessary configuration for the Engine
type Conf struct {
    Recorder     recorder.DataRecorder // The target repository
    IndexName    string
    TypeName     string
    Target       string
    Interval     time.Duration
    Timeout      time.Duration
    Logger       logrus.FieldLogger
    TargetReader reader.TargetReader
}
