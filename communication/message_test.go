// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package communication

import (
	"context"
	"strings"
	"testing"
)

func TestContextKey(t *testing.T) {
	msg := "the key"
	key := contextKey(msg)
	if !strings.Contains(key.String(), msg) {
		t.Errorf("want %s in the key, got (%s)", msg, key.String())
	}
}

func TestNewReadJob(t *testing.T) {
	ctx := context.Background()
	job := NewReadJob(ctx)
	jobID, ok := job.Value(messageID).(JobID)
	if !ok {
		t.Fatalf("want type of JobID, got (%v)", job.Value(messageID))
	}
	if jobID != JobValue(job) {
		t.Errorf("want (%s), got (%s)", jobID, JobValue(job))
	}
	switch job.Value(messageID).(type) {
	case JobID:
		if jobID.String() == "" {
			t.Error("job id is empty")
		}
	default:
		t.Errorf("want JobID type, got (%v)", jobID)
	}
}
