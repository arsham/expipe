// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/arsham/expipe/recorder"
	"github.com/arsham/expipe/tools"
)

// Constructor is an interface for setting up an object for testing.
// TestServer() should return a ready to use test server
type Constructor interface {
	recorder.Constructor
	ValidEndpoints() []string
	InvalidEndpoints() []string
	TestServer() *httptest.Server
	Object() (recorder.DataRecorder, error)
}

// TestSuites returns a map of test name to the runner function.
func TestSuites(t *testing.T, setup func() (Constructor, func())) {
	t.Parallel()
	t.Run("Construction", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		shouldNotChangeTheInput(t, cons)
	})
	t.Run("NameCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		nameCheck(t, cons)
	})
	t.Run("IndexNameCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		indexNameCheck(t, cons)
	})
	t.Run("TimeoutCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		timeoutCheck(t, cons)
	})
	t.Run("EndpointCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		endpointCheck(t, cons)
	})
	t.Run("ReceivesPayload", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		recorderReceivesPayload(t, cons)
	})
	t.Run("SendsResult", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		recorderSendsResult(t, cons)
	})
	t.Run("ErrorsOnUnavailableESServer", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		recorderErrorsOnUnavailableEndpoint(t, cons)
	})
	t.Run("RecordingReturnsErrorIfNotPingedYet", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		recordingReturnsErrorIfNotPingedYet(t, cons)
	})
	t.Run("ErrorsOnBadPayload", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		errorsOnBadPayload(t, cons)
	})
	t.Run("PingingEndpoint", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		pingingEndpoint(t, cons)
	})
	t.Run("ErrorRecordingUnavailableEndpoint", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		errorRecordingUnavailableEndpoint(t, cons)
	})
}

// BaseConstruct implements Constructor interface. It only remembers the setter
// functions, therefore you need to apply them when creating an object in the
// derived constructor. It is concurrent safe.
type BaseConstruct struct {
	sync.Mutex
	setters map[string]func(recorder.Constructor) error
}

// NewBaseConstruct returns an instance of BaseConstruct.
func NewBaseConstruct() *BaseConstruct {
	return &BaseConstruct{
		setters: make(map[string]func(recorder.Constructor) error, 20),
	}
}

// add adds the f function. It will replace the old f if it was called twice.
func (b *BaseConstruct) add(name string, f func(recorder.Constructor) error) {
	b.Lock()
	defer b.Unlock()
	b.setters[name] = f
}

// Setters returns a copy of the configuration functions.
func (b *BaseConstruct) Setters() []func(recorder.Constructor) error {
	b.Lock()
	defer b.Unlock()
	setters := make([]func(recorder.Constructor) error, 0, len(b.setters))
	for _, fn := range b.setters {
		setters = append(setters, fn)
	}

	return setters
}

// SetLogger adds a Logger value to setter configuration.
func (b *BaseConstruct) SetLogger(logger tools.FieldLogger) {
	b.add("logger", recorder.WithLogger(logger))
}

// SetName adds a Name value to setter configuration.
func (b *BaseConstruct) SetName(name string) { b.add("name", recorder.WithName(name)) }

// SetEndpoint adds a Endpoint value to setter configuration.
func (b *BaseConstruct) SetEndpoint(endpoint string) {
	b.add("endpoint", recorder.WithEndpoint(endpoint))
}

// SetIndexName adds an IndexName value to setter configuration.
func (b *BaseConstruct) SetIndexName(indexName string) {
	b.add("indexName", recorder.WithIndexName(indexName))
}

// SetTimeout adds a Timeout value to setter configuration.
func (b *BaseConstruct) SetTimeout(timeout time.Duration) {
	b.add("timeout", recorder.WithTimeout(timeout))
}
