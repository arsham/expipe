// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.
//
// Please note that this file contains tests having logrus inspections.

package expipe

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/reader"
	rt "github.com/arsham/expipe/reader/testing"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

type errMessage string

func (e errMessage) Error() string { return string(e) }

// inspectLogs checks if the niddle is found in the entries.
// The entries might have been stacked, we need to iterate over.
func inspectLogs(entries []*logrus.Entry, niddle string) (all string, found bool) {
	var res []string
	for _, field := range entries {
		if strings.Contains(field.Message, niddle) {
			return "", true
		}
		res = append(res, field.Message)
	}
	return strings.Join(res, ", "), false
}

func TestEventLoopClosingContext(t *testing.T) {
	log, hook := test.NewNullLogger()
	log.Level = internal.DebugLevel

	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil)", err)
	}
	red, err := rt.New(
		reader.WithLogger(internal.DiscardLogger()),
		reader.WithEndpoint(testServer.URL),
		reader.WithName("reader_name"),
		reader.WithTypeName("typeName"),
		reader.WithInterval(time.Hour),
		reader.WithTimeout(time.Hour),
		reader.WithBackoff(5),
	)
	if err != nil {
		t.Fatalf("err = (%#v); want (nil): unexpected error occurred during reader creation", err)
	}
	e.setReaders(map[string]reader.DataReader{red.Name(): red})

	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("expected the engine to quit gracefully")
	}
	if _, found := inspectLogs(hook.Entries, contextCanceled); !found {
		// sometimes it takes time for logrus to register the error, trying again
		time.Sleep(500 * time.Millisecond)
		if all, found := inspectLogs(hook.Entries, contextCanceled); !found {
			t.Errorf("inspectLogs: all = (%v); want (%s) in the error", all, contextCanceled)
		}
	}
}
