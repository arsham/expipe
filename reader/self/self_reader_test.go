// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package self_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/reader/self"
	rt "github.com/arsham/expipe/reader/testing"
)

func getTestServer() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)
}

type Construct struct {
	*self.Reader
	testServer *httptest.Server
}

func (c *Construct) TestServer() *httptest.Server {
	c.testServer = getTestServer()
	return c.testServer
}

func (c *Construct) Object() (reader.DataReader, error) {
	red, err := self.New(
		reader.WithEndpoint(c.Endpoint()),
		reader.WithMapper(datatype.DefaultMapper()),
		reader.WithName(c.Name()),
		reader.WithTypeName(c.TypeName()),
		reader.WithInterval(c.Interval()),
		reader.WithTimeout(c.Timeout()),
		reader.WithBackoff(c.Backoff()),
		reader.WithLogger(internal.DiscardLogger()),
	)
	if err == nil { // FIXME: [refactor] this logic.
		red.SetTestMode()
	}
	return red, err
}

func TestSelfReader(t *testing.T) {
	rt.TestSuites(t, func() (rt.Constructor, func()) {
		r, err := self.New(reader.WithName("test"))
		if err != nil {
			panic(err)
		}
		c := &Construct{Reader: r, testServer: getTestServer()}
		return c, func() { c.testServer.Close() }
	})
}
