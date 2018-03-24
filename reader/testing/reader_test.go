// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/arsham/expipe/reader"
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
	*reader_test.Reader
}

func (c *Construct) TestServer() *httptest.Server { return testServer }
func (c *Construct) Object() (reader.DataReader, error) {
	return reader_test.New(
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

func TestSimpleReader(t *testing.T) {
	r, err := reader_test.New(reader.WithName("test"))
	if err != nil {
		panic(err)
	}
	c := &Construct{r}
	reader_test.TestSuites(t, c)
}
