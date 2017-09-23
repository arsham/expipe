// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/arsham/expipe/recorder"
	recorder_test "github.com/arsham/expipe/recorder/testing"
)

// The purpose of these tests is to make sure the simple recorder, which is a mock,
// works perfect, so other tests can rely on it.

var (
	testServer *httptest.Server
)

func TestMain(m *testing.M) {
	testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	exitCode := m.Run()
	testServer.Close()
	os.Exit(exitCode)
}

func getTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
}

type Construct struct {
	*recorder_test.Recorder
	testServer *httptest.Server
}

func (c *Construct) TestServer() *httptest.Server { return c.testServer }
func (c *Construct) Object() (recorder.DataRecorder, error) {
	return recorder_test.New(
		recorder.SetEndpoint(c.Endpoint()),
		recorder.SetName(c.Name()),
		recorder.SetIndexName(c.IndexName()),
		recorder.SetTimeout(c.Timeout()),
		recorder.SetBackoff(c.Backoff()),
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
	for name, fn := range recorder_test.TestSuites() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			r, err := recorder_test.New()
			if err != nil {
				panic(err)
			}
			c := &Construct{r, getTestServer()}
			defer c.testServer.Close()
			fn(t, c)
		})
	}
}
