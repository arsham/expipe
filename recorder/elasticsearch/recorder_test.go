// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package elasticsearch_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arsham/expipe/recorder"
	"github.com/arsham/expipe/recorder/elasticsearch"
	rt "github.com/arsham/expipe/recorder/testing"
)

func getTestServer() *httptest.Server {
	var host, url, port string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_nodes/http" {
			// sniffing
			w.Write([]byte(fmt.Sprintf(`{
  "_nodes": {
    "total": 1,
    "successful": 1,
    "failed": 0
  },
  "cluster_name": "elasticsearch",
  "nodes": {
    "P2CLuttqTw-UaiqMYwEkeA": {
      "name": "P2CLutt",
      "transport_address": "%s:9300",
      "host": "%s",
      "ip": "%s",
      "version": "5.0.1",
      "build_hash": "080bb47",
      "roles": [
        "master",
        "data",
        "ingest"
      ],
      "http": {
        "bound_address": [
          "[::]:%s"
        ],
        "publish_address": "%s",
        "max_content_length_in_bytes": 104857600
      }
    }
  }
}`, host, host, host, port, url)))
		} else if len(r.URL.Path) > 5 {
			// recording
			w.Write([]byte(`{"_index":"my_index","_type":"my type","_id":"AVlzOSs-sx0uWYTCQCzC","_version":1,"result":"created","_shards":{"total":2,"successful":1,"failed":0},"created":true}`))
		} else if r.URL.Path == "/" {
			// pinging
			w.Write([]byte(`{"name" : "P2CLutt", "cluster_name" : "elasticsearch", "cluster_uuid" : "MEhShuk2R9aUgnnX_Qk2bw", "version" : {"number" : "5.0.1", "build_hash" : "080bb47", "build_date" : "2016-11-11T22:08:49.812Z", "build_snapshot" : false, "lucene_version" : "6.2.1"}, "tagline" : "You Know, for Search"}`))
		}
	})

	testServer := httptest.NewServer(handler)
	url = strings.Split(testServer.URL, "//")[1]
	host, port = strings.Split(url, ":")[0], strings.Split(url, ":")[1]
	return testServer
}

type Construct struct {
	*rt.BaseConstruct
	testServer *httptest.Server
}

func (c *Construct) TestServer() *httptest.Server {
	c.testServer = getTestServer()
	return c.testServer
}

func (c *Construct) Object() (recorder.DataRecorder, error) {
	return elasticsearch.New(c.Setters()...)
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

func TestElasticsearchRecorder(t *testing.T) {
	rt.TestSuites(t, func() (rt.Constructor, func()) {
		c := &Construct{
			testServer:    getTestServer(),
			BaseConstruct: rt.NewBaseConstruct(),
		}
		return c, func() { c.testServer.Close() }
	})
}
