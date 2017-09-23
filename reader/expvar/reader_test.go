// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvar_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/reader/expvar"
	reader_test "github.com/arsham/expipe/reader/testing"
)

var (
	testServer *httptest.Server
)

func TestMain(m *testing.M) {
	testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	exitCode := m.Run()
	testServer.Close()
	os.Exit(exitCode)
}

type Construct struct {
	*expvar.Reader
}

func (c *Construct) TestServer() *httptest.Server { return testServer }
func (c *Construct) Object() (reader.DataReader, error) {
	return expvar.New(
		reader.SetEndpoint(c.Endpoint()),
		reader.SetMapper(c.Mapper()),
		reader.SetName(c.Name()),
		reader.SetTypeName(c.TypeName()),
		reader.SetInterval(c.Interval()),
		reader.SetTimeout(c.Timeout()),
		reader.SetBackoff(c.Backoff()),
	)
}

func TestExpvar(t *testing.T) {
	for name, fn := range reader_test.TestSuites() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			r, err := expvar.New(reader.SetName("test"))
			if err != nil {
				panic(err)
			}
			fn(t, &Construct{r})
		})
	}
}
