// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arsham/expipe/reader"
	rt "github.com/arsham/expipe/reader/testing"
	"github.com/arsham/expipe/token"
	"github.com/pkg/errors"
)

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

func (c *Construct) Object() (reader.DataReader, error) {
	return rt.New(c.Setters()...)
}

func TestMockReader(t *testing.T) {
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
	red := rt.Reader{
		PingFunc: func() error {
			if !called {
				called = true
				return err1
			} else {
				return err2
			}
		},
	}
	err := red.Ping()
	if !called {
		t.Error("Ping(): called = (false); want (true)")
	}
	if err != err1 {
		t.Errorf("Ping() = (%#v); want (%v)", err, err1)
	}
	err = red.Ping()
	if err != err2 {
		t.Errorf("Ping() = (%#v); want (%v)", err, err2)
	}
}

func TestReadFunc(t *testing.T) {
	var (
		called bool
		err1   = errors.New("error 1")
		err2   = errors.New("error 2")
	)
	red := rt.Reader{
		ReadFunc: func(*token.Context) (*reader.Result, error) {
			if !called {
				called = true
				return nil, err1
			} else {
				return nil, err2
			}
		},
	}
	_, err := red.Read(nil)
	if !called {
		t.Error("Read(nil): called = (false); want (true)")
	}
	if err != err1 {
		t.Errorf("Read(nil) = (%#v); want (%v)", err, err1)
	}
	_, err = red.Read(nil)
	if err != err2 {
		t.Errorf("Read(nil) = (%#v); want (%v)", err, err2)
	}
}
