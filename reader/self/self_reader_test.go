// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package self_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/internal/datatype"
	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/reader/self"
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
	*self.Reader
}

func (c *Construct) TestServer() *httptest.Server { return testServer }
func (c *Construct) Object() (reader.DataReader, error) {
	red, err := self.New(
		reader.SetEndpoint(c.Endpoint()),
		reader.SetMapper(datatype.DefaultMapper()),
		reader.SetName(c.Name()),
		reader.SetTypeName(c.TypeName()),
		reader.SetInterval(c.Interval()),
		reader.SetTimeout(c.Timeout()),
		reader.SetBackoff(c.Backoff()),
		reader.SetLogger(internal.DiscardLogger()),
	)
	if err == nil {
		red.SetTestMode()
	}
	return red, err
}

func TestSelfReader(t *testing.T) {
	r, err := self.New(reader.SetName("test"))
	if err != nil {
		panic(err)
	}
	c := &Construct{r}
	reader_testing.TestSuites(t, c)
}
