// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.
//
// +build !race
//
// Please note that this file contains tests having logrus inspections.

package expipe

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/test"
	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/internal/token"
	"github.com/arsham/expipe/reader"
	reader_test "github.com/arsham/expipe/reader/testing"
)

type errMessage string

func (e errMessage) Error() string { return string(e) }

// inspectLogs checks if the niddle is found in the entries
// the entries might have been stacked, we need to iterate over.
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

func TestEventLoopCatchesReaderError(t *testing.T) {
	log, hook := test.NewNullLogger()
	log.Level = internal.ErrorLevel

	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	red, err := reader_test.New(
		reader.SetLogger(internal.DiscardLogger()),
		reader.SetEndpoint(testServer.URL),
		reader.SetName("reader_name"),
		reader.SetTypeName("typeName"),
		reader.SetInterval(10*time.Millisecond),
		reader.SetTimeout(time.Second),
		reader.SetBackoff(5),
	)
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}
	if err = red.Ping(); err != nil {
		t.Fatal(err)
	}

	e.setReaders(map[string]reader.DataReader{red.Name(): red})

	errMsg := errMessage("an error happened")
	recorded := make(chan struct{})

	// Testing the engine catches errors
	red.ReadFunc = func(job *token.Context) (*reader.Result, error) {
		recorded <- struct{}{}
		return nil, errMsg
	}

	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()

	select {
	case <-recorded:
	case <-time.After(5 * time.Second):
		t.Error("expected to record, didn't happen")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}

	if _, found := inspectLogs(hook.Entries, errMsg.Error()); !found {
		// sometimes it takes time for logrus to register the error, trying again
		time.Sleep(500 * time.Millisecond)
		if all, found := inspectLogs(hook.Entries, errMsg.Error()); !found {
			t.Errorf("want (%s) in the error, got (%v)", errMsg.Error(), all)
		}
	}
}

func TestEventLoopClosingContext(t *testing.T) {
	log, hook := test.NewNullLogger()
	log.Level = internal.DebugLevel

	ctx, cancel := context.WithCancel(context.Background())
	e, err := withRecorder(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	red, err := reader_test.New(
		reader.SetLogger(internal.DiscardLogger()),
		reader.SetEndpoint(testServer.URL),
		reader.SetName("reader_name"),
		reader.SetTypeName("typeName"),
		reader.SetInterval(time.Hour),
		reader.SetTimeout(time.Hour),
		reader.SetBackoff(5),
	)
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
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
			t.Errorf("want (%s) in the error, got (%v)", contextCanceled, all)
		}
	}
}
