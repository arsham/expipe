// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arsham/expipe/recorder"
	rt "github.com/arsham/expipe/recorder/testing"
	"github.com/pkg/errors"
)

// The purpose of these tests is to make sure the simple recorder, which is
// a mock, works perfect, so other tests can rely on it.

func getTestServer() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)
}

type Construct struct {
	*rt.BaseConstruct
	testServer *httptest.Server
}

func (c *Construct) TestServer() *httptest.Server {
	c.testServer = getTestServer()
	return c.testServer
}

func (c *Construct) Object() (recorder.DataRecorder, error) {
	return rt.New(c.Setters()...)
}

func (c *Construct) ValidEndpoints() []string {
	return []string{
		"http://192.168.1.1:9200",
		"http://127.0.0.1:9200",
		"http://localhost:9200",
		"http://localhost.localdomain:9200",
	}
}

func (c *Construct) InvalidEndpoints() []string {
	return []string{
		"http://192.168 .1.1:9200",
		"http ://127.0.0.1:9200",
		"http://:9200",
		":9200",
		"",
	}
}

func TestMockRecorder(t *testing.T) {
	rt.TestSuites(t, func() (rt.Constructor, func()) {
		c := &Construct{
			testServer:    getTestServer(),
			BaseConstruct: rt.NewBaseConstruct(),
		}
		return c, func() { c.testServer.Close() }
	})
}

func TestPingFunc(t *testing.T) {
	var (
		called bool
		err1   = errors.New("error 1")
		err2   = errors.New("error 2")
	)
	rec := rt.Recorder{
		PingFunc: func() error {
			if !called {
				called = true
				return err1
			}
			return err2
		},
	}
	err := rec.Ping()
	if !called {
		t.Error("Ping(): called = (false); want (true)")
	}
	if err != err1 {
		t.Errorf("Ping() = (%#v); want (%v)", err, err1)
	}
	err = rec.Ping()
	if err != err2 {
		t.Errorf("Ping() = (%#v); want (%v)", err, err2)
	}
}

func TestPinged(t *testing.T) {
	called := false
	err1 := errors.New("error 1")
	rec := rt.Recorder{}
	rec.PingFunc = func() error {
		if !called {
			called = true
			rec.Pinged = true
			return err1
		}
		t.Error("didn't expect the ping barrier to pass")
		return nil
	}

	err2 := rec.Ping()
	if err2 != err1 {
		t.Errorf("Ping() = (%#v); want (%v)", err2, err1)
	}
	if !called {
		t.Error("Ping(): called = (false); want (true)")
	}
	err2 = rec.Ping()
	if err2 != nil {
		t.Errorf("Ping() = (%#v); want (nil)", err2)
	}

}

func TestReadFunc(t *testing.T) {
	var (
		called bool
		err1   = errors.New("error 1")
		err2   = errors.New("error 2")
	)
	rec := rt.Recorder{
		RecordFunc: func(context.Context, *recorder.Job) error {
			if !called {
				called = true
				return err1
			} else {
				return err2
			}
		},
	}
	err := rec.Record(context.TODO(), nil)
	if !called {
		t.Error("Record(nil): called = (false); want (true)")
	}
	if err != err1 {
		t.Errorf("Record(nil) = (%#v); want (%v)", err, err1)
	}
	err = rec.Record(context.TODO(), nil)
	if err != err2 {
		t.Errorf("Record(nil) = (%#v); want (%v)", err, err2)
	}
}
