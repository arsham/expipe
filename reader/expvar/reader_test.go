// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvar_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/reader/expvar"
	rt "github.com/arsham/expipe/reader/testing"
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
	return expvar.New(c.Setters()...)
}

func TestExpvarReader(t *testing.T) {
	rt.TestSuites(t, func() (rt.Constructor, func()) {
		c := &Construct{
			testServer:    getTestServer(),
			BaseConstruct: rt.NewBaseConstruct(),
		}
		return c, func() { c.testServer.Close() }
	})
}
