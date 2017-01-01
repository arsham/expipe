// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package self

import (
	"context"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/lib"
)

// The other test goes through a normal path, we need to test the actual path
func TestSelfReaderReadsExpvar(t *testing.T) {
	log := lib.DiscardLogger()
	typeName := "my type"
	mapper := datatype.DefaultMapper()
	red := &Reader{
		name:     "self",
		typeName: typeName,
		mapper:   mapper,
		log:      log,
		interval: time.Hour,
		timeout:  time.Hour,
		endpoint: IgnoredEndpoint,
		backoff:  5,
	}
	job := communication.NewReadJob(context.Background())
	res, err := red.Read(job)
	if err != nil {
		t.Fatalf("want nil, got (%s)", err)
	}
	if res == nil {
		t.Fatal("want result, got nil")
	}
	if res.ID != communication.JobValue(job) {
		t.Errorf("want (%s), got (%s)", res.ID, communication.JobValue(job))
	}
	if res.TypeName != typeName {
		t.Errorf("want (%s), got (%s)", typeName, res.TypeName)
	}
	if res.Mapper != mapper {
		t.Errorf("want (%s), got (%s)", typeName, res.TypeName)
	}

	container := datatype.JobResultDataTypes(res.Res, mapper)
	if container.Len() == 0 {
		t.Error("empty container")
	}
}
