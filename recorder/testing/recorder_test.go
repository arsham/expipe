// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/recorder"
	recorder_test "github.com/arsham/expvastic/recorder/testing"
)

// The purpose of these tests is to make sure the simple recorder, which is a mock,
// works perfect, so other tests can rely on it.

var (
	log        logrus.FieldLogger
	testServer *httptest.Server
)

func TestMain(m *testing.M) {
	log = lib.DiscardLogger()
	testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	exitCode := m.Run()
	testServer.Close()
	os.Exit(exitCode)
}

type Construct struct {
	name      string
	indexName string
	endpoint  string
	interval  time.Duration
	timeout   time.Duration
	backoff   int
}

func (c *Construct) SetName(name string)                { c.name = name }
func (c *Construct) SetIndexName(indexName string)      { c.indexName = indexName }
func (c *Construct) SetEndpoint(endpoint string)        { c.endpoint = endpoint }
func (c *Construct) SetInterval(interval time.Duration) { c.interval = interval }
func (c *Construct) SetTimeout(timeout time.Duration)   { c.timeout = timeout }
func (c *Construct) SetBackoff(backoff int)             { c.backoff = backoff }
func (c *Construct) TestServer() *httptest.Server       { return testServer }
func (c *Construct) Object() (recorder.DataRecorder, error) {
	return recorder_test.New(context.Background(), log, c.name, c.endpoint, c.indexName, c.timeout, c.backoff)
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
	recorder_test.TestRecorder(t, &Construct{})
}
