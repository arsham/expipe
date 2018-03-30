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
)

func getTestServer() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)
}

type Construct struct {
	*rt.Reader
	testServer *httptest.Server
}

func (c *Construct) TestServer() *httptest.Server {
	c.testServer = getTestServer()
	return c.testServer
}

func (c *Construct) Object() (reader.DataReader, error) {
	return rt.New(
		reader.WithEndpoint(c.Endpoint()),
		reader.WithName(c.Name()),
		reader.WithTypeName(c.TypeName()),
		reader.WithInterval(c.Interval()),
		reader.WithTimeout(c.Timeout()),
		reader.WithBackoff(c.Backoff()),
		reader.WithMapper(c.Mapper()),
		reader.WithLogger(c.Logger()),
	)
}

func TestMockReader(t *testing.T) {
	rt.TestSuites(t, func() (rt.Constructor, func()) {
		r, err := rt.New(reader.WithName("test"))
		if err != nil {
			panic(err)
		}
		c := &Construct{Reader: r, testServer: getTestServer()}
		return c, func() { c.testServer.Close() }
	})
}
