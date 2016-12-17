// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

import (
    "context"
    "time"

    "github.com/arsham/expvastic"
)

type mockRecorder struct {
    RecordFunc func(ctx context.Context, typeName string, t time.Time, list expvastic.DataContainer) error
}

func (m *mockRecorder) Record(ctx context.Context, typeName string, t time.Time, list expvastic.DataContainer) error {
    if m.RecordFunc != nil {
        return m.RecordFunc(ctx, typeName, t, list)
    }
    return nil
}
