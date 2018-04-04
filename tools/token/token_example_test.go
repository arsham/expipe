// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package token_test

import (
	"context"
	"testing"

	"github.com/arsham/expipe/tools/token"
)

func ExampleContext_new(t *testing.T) {
	ctx := context.Background()
	token.New(ctx)
}

// func TestNewReadJob(t *testing.T) {
// 	ctx := context.Background()
// 	job := New(ctx)
// 	jobID, ok := job.Value(tokenID).(ID)
// 	if !ok {
// 		t.Fatalf("want type of JobID, got (%v)", job.Value(tokenID))
// 	}
// 	if jobID != job.ID() {
// 		t.Errorf("want (%s), got (%s)", jobID, job.ID())
// 	}
// 	switch job.Value(tokenID).(type) {
// 	case ID:
// 		if jobID.String() == "" {
// 			t.Error("job id is empty")
// 		}
// 	default:
// 		t.Errorf("want JobID type, got (%v)", jobID)
// 	}
// }
