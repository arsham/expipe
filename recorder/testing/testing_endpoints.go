// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/arsham/expipe/internal/datatype"
	"github.com/arsham/expipe/internal/token"
	"github.com/arsham/expipe/recorder"
	gin "github.com/onsi/ginkgo"
	gom "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

// testRecorderErrorsOnUnavailableEndpoint tests the recorder errors for bad URL.
func testRecorderErrorsOnUnavailableEndpoint(cons Constructor) {
	gin.Context("by initiated recorder pointing to an unavailable server", func() {
		var (
			err error
			rec recorder.DataRecorder
		)
		timeout := time.Second
		name := "the name"
		indexName := "my_index_name"
		backoff := 5
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer ts.Close()
		cons.SetName(name)
		cons.SetIndexName(indexName)
		cons.SetEndpoint(ts.URL)
		cons.SetTimeout(timeout)
		cons.SetBackoff(backoff)
		ts.Close()

		gin.Context("when getting the object", func() {
			rec, err = cons.Object()
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
		})

		gin.Context("when pinging", func() {
			err := rec.Ping()
			gin.It("should error", func() {
				gom.Expect(err).To(gom.HaveOccurred())
				err = errors.Cause(err)
				gom.Expect(err).To(gom.BeAssignableToTypeOf(recorder.ErrEndpointNotAvailable{}))
			})
		})

	})
}

// testRecorderBacksOffOnEndpointGone is a helper to test the recorder backs off when the endpoint goes away.
func testRecorderBacksOffOnEndpointGone(cons Constructor) {
	gin.Context("by initiating a recorder and having a running endpoint", func() {
		var (
			rec recorder.DataRecorder
			err error
		)
		ctx := context.Background()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer ts.Close()
		timeout := time.Second
		cons.SetName("the name")
		cons.SetIndexName("my_index_name")
		cons.SetEndpoint(ts.URL)
		cons.SetTimeout(timeout)
		cons.SetBackoff(5)

		gin.Context("when getting the recorder object", func() {
			rec, err = cons.Object()
			gin.It("should return a recorder", func() {
				gom.Expect(rec).NotTo(gom.BeNil())
			})
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
		})

		gin.Context("when pinging", func() {
			err := rec.Ping()
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
		})

		gin.Context("when closing the server and having a payload to be sent", func() {
			ts.Close()

			p := datatype.New([]datatype.DataType{})
			payload := &recorder.Job{
				ID:        token.NewUID(),
				Payload:   p,
				IndexName: "my index",
				TypeName:  "my type",
				Time:      time.Now(),
			}
			gin.Context("while draining the recorder", func() {

				// We don't know the backoff amount set in the recorder, so we try 100 times until it closes.
				backedOff := false
				for i := 0; i < 100; i++ {
					err := rec.Record(ctx, payload)
					if err == recorder.ErrBackoffExceeded {
						backedOff = true
						break
					}
				}
				gin.It("should exceed the backoff", func() {
					gom.Expect(backedOff).To(gom.BeTrue())
				})
			})

			gin.Context("sending another payload", func() {

				gin.It("should block", func(done gin.Done) {
					stop := make(chan struct{})
					go func() {
						rec.Record(ctx, payload)
						close(stop)
					}()
					<-stop
					close(done)
				}, 0.02)
			})
		})
	})
}

// testRecordingReturnsErrorIfNotPingedYet is a helper to test the recorder returns an error
// if the caller hasn't called the Ping() method.
func testRecordingReturnsErrorIfNotPingedYet(cons Constructor) {
	gin.Context("With a recorder initialised", func() {
		var (
			rec recorder.DataRecorder
			err error
		)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer ts.Close()
		ctx := context.Background()
		timeout := time.Second
		cons.SetName("the name")
		cons.SetIndexName("my_index_name")
		cons.SetTimeout(timeout)
		cons.SetEndpoint(ts.URL)
		cons.SetBackoff(5)

		gin.Context("when getting the recorder object", func() {
			rec, err = cons.Object()
			gin.It("should return a recorder", func() {
				gom.Expect(rec).NotTo(gom.BeNil())
			})
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
		})

		gin.Context("having a payload ready, recording without pinging", func() {
			p := datatype.New([]datatype.DataType{})
			payload := &recorder.Job{
				ID:        token.NewUID(),
				Payload:   p,
				IndexName: "my index",
				TypeName:  "my type",
				Time:      time.Now(),
			}
			err := rec.Record(ctx, payload)
			gin.Specify("should error", func() {
				gom.Expect(errors.Cause(err)).To(gom.MatchError(recorder.ErrPingNotCalled))
			})
		})
	})
}
