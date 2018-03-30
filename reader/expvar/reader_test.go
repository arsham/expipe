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
	*expvar.Reader
	testServer *httptest.Server
}

func (c *Construct) TestServer() *httptest.Server {
	c.testServer = getTestServer()
	return c.testServer
}

func (c *Construct) Object() (reader.DataReader, error) {
	return expvar.New(
		reader.WithEndpoint(c.Endpoint()),
		reader.WithMapper(c.Mapper()),
		reader.WithName(c.Name()),
		reader.WithTypeName(c.TypeName()),
		reader.WithInterval(c.Interval()),
		reader.WithTimeout(c.Timeout()),
		reader.WithBackoff(c.Backoff()),
	)
}

func TestExpvarReader(t *testing.T) {
	rt.TestSuites(t, func() (rt.Constructor, func()) {
		r, err := expvar.New(reader.WithName("test"))
		if err != nil {
			panic(err)
		}
		c := &Construct{Reader: r, testServer: getTestServer()}
		return c, func() { c.testServer.Close() }
	})
}
