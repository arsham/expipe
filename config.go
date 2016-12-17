// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
    "context"
    "time"

    "github.com/Sirupsen/logrus"
)

// Recorder in an interface for shipping data to a repository.
// The repository should have the concept of type/table. Which is inside index/database abstraction. See ElasticSearch for more information.
type Recorder interface {
    // timestamp is used for timeseries data
    Record(ctx context.Context, typeName string, timestamp time.Time, list DataContainer) error
}

type targetReader interface {
    JobChan() chan context.Context
    ResultChan() chan JobResult
}

// Conf holds the necessary configuration for the Client
type Conf struct {
    Recorder     Recorder // The target repository
    IndexName    string
    TypeName     string
    Target       string
    Interval     time.Duration
    Timeout      time.Duration
    Logger       logrus.FieldLogger
    TargetReader targetReader
}
