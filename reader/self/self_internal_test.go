// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package self

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/token"
)

// The other test goes through a normal path, we need to test the actual path.
func TestSelfReaderReadsExpvar(t *testing.T) {
	log := internal.DiscardLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	typeName := "my_type"
	mapper := datatype.DefaultMapper()
	red := &Reader{
		name:     "self",
		typeName: typeName,
		mapper:   mapper,
		log:      log,
		interval: time.Hour,
		timeout:  time.Hour,
		endpoint: ts.URL,
		backoff:  5,
		testMode: true, // so we can ping, then we will make it false.
	}
	err := red.Ping()
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red.testMode = false // set it so it goes through the normal mode.
	job := token.New(context.Background())
	res, err := red.Read(job)
	if err != nil {
		t.Fatalf("err = (%s); want (nil)", err)
	}
	if res == nil {
		t.Fatal("res = (nil); want (result)")
	}
	if res.ID != job.ID() {
		t.Errorf("res.ID = (%s); want (%s)", job.ID(), res.ID)
	}
	if res.TypeName != typeName {
		t.Errorf("res.TypeName = (%s); want (%s)", res.TypeName, typeName)
	}
	if res.Mapper != mapper {
		t.Errorf("res.TypeName = (%s); want (%s)", res.TypeName, typeName)
	}
	container, _ := datatype.JobResultDataTypes(res.Content, mapper)
	if container.Len() == 0 {
		t.Error("container.Len() = 0; want (!= 0)")
	}
}
