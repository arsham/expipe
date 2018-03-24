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
	reader_testing "github.com/arsham/expipe/reader/testing"
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
	r, err := expvar.New(reader.WithName("test"))
	if err != nil {
		panic(err)
	}
	c := &Construct{r}
	reader_testing.TestSuites(t, c)
}
