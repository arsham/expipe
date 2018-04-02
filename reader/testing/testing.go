// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/tools"
)

// Constructor is an interface for setting up an object for testing.
// TestServer should return a ready to use test server
// Object should return the instantiated object
type Constructor interface {
	reader.Constructor
	TestServer() *httptest.Server
	Object() (reader.DataReader, error)
}

// TestSuites returns a map of test name to the runner function.
func TestSuites(t *testing.T, setup func() (Constructor, func())) {
	t.Parallel()
	t.Run("ShouldNotChangeTheInput", func(t *testing.T) {
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
	t.Run("TypeNameCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		typeNameCheck(t, cons)
	})
	t.Run("BackoffCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		backoffCheck(t, cons)
	})
	t.Run("IntervalCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		intervalCheck(t, cons)
	})
	t.Run("EndpointCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		endpointCheck(t, cons)
	})
	t.Run("TimeoutCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		timeoutCheck(t, cons)
	})
	t.Run("SetMapperCheck", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		setMapperCheck(t, cons)
	})
	t.Run("ReceivesJob", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		readerReceivesJob(t, cons)
	})
	t.Run("ReturnsSameID", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		readerReturnsSameID(t, cons)
	})
	t.Run("PingingEndpoint", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		pingingEndpoint(t, cons)
	})
	t.Run("ErrorsOnEndpointDisapears", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		readerErrorsOnEndpointDisapears(t, cons)
	})
	t.Run("BacksOffOnEndpointGone", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		readerBacksOffOnEndpointGone(t, cons)
	})
	t.Run("ReadingReturnsErrorIfNotPingedYet", func(t *testing.T) {
		t.Parallel()
		cons, cleanup := setup()
		defer cleanup()
		readingReturnsErrorIfNotPingedYet(t, cons)
	})
}

// BaseConstruct implements Constructor interface.
// It only remembers the setter functions, therefore you need to apply them
// when creating an object in the derived constructor. It is concurrent safe.
type BaseConstruct struct {
	sync.Mutex
	setters map[string]func(reader.Constructor) error
}

// NewBaseConstruct returns an instance of BaseConstruct.
func NewBaseConstruct() *BaseConstruct {
	return &BaseConstruct{
		setters: make(map[string]func(reader.Constructor) error, 20),
	}
}

// add adds the f function. It will replace the old f if it was called twice.
func (b *BaseConstruct) add(name string, f func(reader.Constructor) error) {
	b.Lock()
	defer b.Unlock()
	b.setters[name] = f
}

// Setters returns a copy of the configuration functions.
func (b *BaseConstruct) Setters() []func(reader.Constructor) error {
	b.Lock()
	defer b.Unlock()
	setters := make([]func(reader.Constructor) error, 0, len(b.setters))
	for _, fn := range b.setters {
		setters = append(setters, fn)
	}

	return setters
}

// SetLogger adds a Logger value to setter configuration.
func (b *BaseConstruct) SetLogger(logger tools.FieldLogger) {
	b.add("logger", reader.WithLogger(logger))
}

// SetName adds a Name value to setter configuration.
func (b *BaseConstruct) SetName(name string) { b.add("name", reader.WithName(name)) }

// SetTypeName adds a TypeName value to setter configuration.
func (b *BaseConstruct) SetTypeName(typeName string) { b.add("typeName", reader.WithTypeName(typeName)) }

// SetEndpoint adds a Endpoint value to setter configuration.
func (b *BaseConstruct) SetEndpoint(endpoint string) { b.add("endpoint", reader.WithEndpoint(endpoint)) }

// SetMapper adds a Mapper value to setter configuration.
func (b *BaseConstruct) SetMapper(mapper datatype.Mapper) { b.add("mapper", reader.WithMapper(mapper)) }

// SetBackoff adds a Backoff value to setter configuration.
func (b *BaseConstruct) SetBackoff(backoff int) { b.add("backoff", reader.WithBackoff(backoff)) }

// SetInterval adds a Interval value to setter configuration.
func (b *BaseConstruct) SetInterval(interval time.Duration) {
	b.add("interval", reader.WithInterval(interval))
}

// SetTimeout adds a Timeout value to setter configuration.
func (b *BaseConstruct) SetTimeout(timeout time.Duration) {
	b.add("timeout", reader.WithTimeout(timeout))
}
