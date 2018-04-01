// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package testing is a test suit for readers. They should provide
// an object that implements the Constructor interface then run:
//
//    import rt "github.com/arsham/expipe/reader/testing"
//    ....
//    type Construct struct {
//        *rt.BaseConstruct
//        testServer *httptest.Server
//    }
//
//    func (c *Construct) TestServer() *httptest.Server {
//        return /* a test server */
//    }
//
//    func (c *Construct) Object() (reader.DataReader, error) {
//        return expvar.New(c.Setters()...)
//    }
//
//    func TestMyReader(t *testing.T) {
//        rt.TestSuites(t, func() (rt.Constructor, func()) {
//            c := &Construct{
//                testServer:    getTestServer(),
//                BaseConstruct: rt.NewBaseConstruct(),
//            }
//            return c, func() { c.testServer.Close() }
//        })
//    }
//
// The test suit will pick it up and does all the tests.
//
// Important Note
//
// The test suite might close and request the test server multiple times during
// its work. Make sure you return a brand new instance every time the TestServer
// method is been called.
// All tests are ran in isolation and they are set to be run as parallel. Make
// sure your code doesn't have any race conditions.
// You need to write the edge cases if they are not covered in this section.
//
package testing
