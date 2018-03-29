// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/token"
	gin "github.com/onsi/ginkgo"
	gom "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

// pingingEndpoint is a helper to test the reader errors when the endpoint goes away.
func pingingEndpoint(cons Constructor) {
	gin.Context("having a reader set up", func() {
		ts := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		ts.Close()
		cons.SetName("the name")
		cons.SetTypeName("my type")
		cons.SetEndpoint(ts.URL)
		cons.SetInterval(time.Millisecond)
		cons.SetTimeout(time.Second)

		gin.Context("when creating the object", func() {
			red, err := cons.Object()
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
			gin.Specify("reader should not be nil", func() {
				gom.Expect(red).NotTo(gom.BeNil())
			})

			gin.Context("when pinging", func() {
				err := red.Ping()
				gin.It("should error", func() {
					gom.Expect(err).To(gom.HaveOccurred())
					gom.Expect(errors.Cause(err)).To(
						gom.BeAssignableToTypeOf(reader.ErrEndpointNotAvailable{}),
					)
				})
			})
		})

		gin.Context("when pointing to an unavailable endpoint", func() {
			unavailableEndpoint := "http://192.168.255.255"
			cons.SetEndpoint(unavailableEndpoint)
			red, _ := cons.Object()
			gin.Context("by pinging", func() {
				err := red.Ping()
				gin.It("should error and mention the endpoint", func() {
					gom.Expect(err).To(gom.HaveOccurred())
					gom.Expect(errors.Cause(err)).To(
						gom.BeAssignableToTypeOf(reader.ErrEndpointNotAvailable{}),
					)
					gom.Expect(err.Error()).To(gom.ContainSubstring(unavailableEndpoint))
				})
			})
		})
	})
}

// testReaderErrorsOnEndpointDisapears is a helper to test the reader errors
// when the endpoint goes away.
func testReaderErrorsOnEndpointDisapears(cons Constructor) {
	gin.Context("having the reader initiated and pointing to a running endpoint", func() {
		var (
			red reader.DataReader
			err error
		)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		cons.SetName("the name")
		cons.SetTypeName("my type")
		cons.SetEndpoint(ts.URL)
		cons.SetInterval(time.Hour)
		cons.SetTimeout(time.Hour)
		cons.SetBackoff(5)

		gin.Context("when creating the object", func() {
			red, err = cons.Object()
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
		})

		gin.Context("when pinging the endpoint", func() {
			err := red.Ping()
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
		})

		gin.Context("having the endpoint server closed", func() {
			ts.Close()

			gin.Context("when reading from the endpoint", func() {
				ctx := context.Background()
				result, err := red.Read(token.New(ctx))
				gin.It("should error and mention the url", func() {
					gom.Expect(err).To(gom.HaveOccurred())
					err = errors.Cause(err)
					gom.Expect(err).To(
						gom.BeAssignableToTypeOf(reader.ErrEndpointNotAvailable{}),
					)
					gom.Expect(err.Error()).To(gom.ContainSubstring(ts.URL))
				})
				gin.Specify("the result should be nil", func() {
					gom.Expect(result).To(gom.BeNil())
				})
			})
		})
	})
}

// testReaderBacksOffOnEndpointGone is a helper to test the reader backs off
// when the endpoint goes away.
func testReaderBacksOffOnEndpointGone(cons Constructor) {
	gin.Context("having the reader initiated and pointing to a running endpoint", func() {
		var (
			red reader.DataReader
			err error
		)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		cons.SetName("the name")
		cons.SetTypeName("my type")
		cons.SetEndpoint(ts.URL)
		cons.SetInterval(time.Hour)
		cons.SetTimeout(time.Hour)
		cons.SetBackoff(5)

		gin.Context("when creating the object", func() {
			red, err = cons.Object()
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
		})

		gin.Context("when pinging the endpoint", func() {
			err := red.Ping()
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
		})

		gin.Context("by closing the server and issuing a read job", func() {
			ts.Close()
			ctx := context.Background()
			job := token.New(ctx)

			backedOff := false
			// We don't know the backoff amount set in the reader, so we try
			// 100 times until it closes.
			for i := 0; i < 100; i++ {
				_, err := red.Read(job)
				if err == reader.ErrBackoffExceeded {
					backedOff = true
					break
				}
			}
			gin.It("should exceed the backoff", func() {
				// FIXME: this test is skipped (0)
				gin.Skip("skipped for now")
				gom.Expect(backedOff).To(gom.BeTrue())
			})

			gin.Context("sending another job, it should block", func() {
				gin.It("should be gone", func() {
					r, err := red.Read(job)
					gom.Expect(err).To(gom.HaveOccurred())
					gom.Expect(r).To(gom.BeNil())
				})
			})
		})

	})
}

// testReadingReturnsErrorIfNotPingedYet is a helper to test the reader
// returns an error if the caller hasn't called the Ping() method.
func testReadingReturnsErrorIfNotPingedYet(cons Constructor) {
	gin.Context("With a reader initialised", func() {
		var (
			red reader.DataReader
			err error
		)

		ctx := context.Background()
		cons.SetName("the name")
		cons.SetTypeName("my type")
		cons.SetEndpoint(cons.TestServer().URL)
		cons.SetInterval(time.Second)
		cons.SetTimeout(time.Second)
		cons.SetBackoff(5)

		gin.Context("when creating the object", func() {
			red, err = cons.Object()
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
		})

		gin.Context("when issuing a read job", func() {
			job := token.New(ctx)
			res, err := red.Read(job)

			gin.It("should error with ErrPingNotCalled", func() {
				gom.Expect(err).To(gom.HaveOccurred())
				gom.Expect(errors.Cause(err)).To(gom.MatchError(reader.ErrPingNotCalled))
			})
			gin.Specify("result should be empty", func() {
				gom.Expect(res).To(gom.BeNil())
			})
		})
	})
}
