// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arsham/expipe/recorder"
	rt "github.com/arsham/expipe/recorder/testing"
)

// The purpose of these tests is to make sure the simple recorder, which is
// a mock, works perfect, so other tests can rely on it.

func getTestServer() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)
}

type Construct struct {
	*rt.Recorder
	testServer *httptest.Server
}

func (c *Construct) TestServer() *httptest.Server {
	c.testServer = getTestServer()
	return c.testServer
}

func (c *Construct) Object() (recorder.DataRecorder, error) {
	return rt.New(
		recorder.WithEndpoint(c.Endpoint()),
		recorder.WithName(c.Name()),
		recorder.WithIndexName(c.IndexName()),
		recorder.WithTimeout(c.Timeout()),
		recorder.WithBackoff(c.Backoff()),
	)
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
		r, err := rt.New()
		if err != nil {
			panic(err)
		}
		c := &Construct{Recorder: r, testServer: getTestServer()}
		return c, func() { c.testServer.Close() }
	})
}
