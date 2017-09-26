// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/arsham/expipe/internal/token"
	"github.com/arsham/expipe/reader"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

// pingingEndpoint is a helper to test the reader errors when the endpoint goes away.
func pingingEndpoint(cons Constructor) {
	Context("having a reader set up", func() {
		ts := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		ts.Close()
		cons.SetName("the name")
		cons.SetTypeName("my type")
		cons.SetEndpoint(ts.URL)
		cons.SetInterval(time.Millisecond)
		cons.SetTimeout(time.Second)

		Context("when creating the object", func() {
			red, err := cons.Object()
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			Specify("reader should not be nil", func() {
				Expect(red).NotTo(BeNil())
			})

			Context("when pinging", func() {
				err := red.Ping()
				It("should error", func() {
					Expect(err).To(HaveOccurred())
					Expect(errors.Cause(err)).To(BeAssignableToTypeOf(reader.ErrEndpointNotAvailable{}))
				})
			})
		})

		Context("when pointing to an unavailable endpoint", func() {
			unavailableEndpoint := "http://192.168.255.255"
			cons.SetEndpoint(unavailableEndpoint)
			red, _ := cons.Object()
			Context("by pinging", func() {
				err := red.Ping()
				It("should error and mention the endpoint", func() {
					Expect(err).To(HaveOccurred())
					Expect(errors.Cause(err)).To(BeAssignableToTypeOf(reader.ErrEndpointNotAvailable{}))
					Expect(err.Error()).To(ContainSubstring(unavailableEndpoint))
				})
			})
		})
	})
}

// testReaderErrorsOnEndpointDisapears is a helper to test the reader errors when the endpoint goes away.
func testReaderErrorsOnEndpointDisapears(cons Constructor) {
	Context("having the reader initiated and pointing to a running endpoint", func() {
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

		Context("when creating the object", func() {
			red, err = cons.Object()
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when pinging the endpoint", func() {
			err := red.Ping()
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("having the endpoint server closed", func() {
			ts.Close()

			Context("when reading from the endpoint", func() {
				ctx := context.Background()
				result, err := red.Read(token.New(ctx))
				It("should error and mention the url", func() {
					Expect(err).To(HaveOccurred())
					err = errors.Cause(err)
					Expect(err).To(BeAssignableToTypeOf(reader.ErrEndpointNotAvailable{}))
					Expect(err.Error()).To(ContainSubstring(ts.URL))
				})
				Specify("the result should be nil", func() {
					Expect(result).To(BeNil())
				})
			})
		})
	})
}

// testReaderBacksOffOnEndpointGone is a helper to test the reader backs off when the endpoint goes away.
func testReaderBacksOffOnEndpointGone(cons Constructor) {
	Context("having the reader initiated and pointing to a running endpoint", func() {
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

		Context("when creating the object", func() {
			red, err = cons.Object()
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when pinging the endpoint", func() {
			err := red.Ping()
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("by closing the server and issuing a read job", func() {
			ts.Close()
			ctx := context.Background()
			job := token.New(ctx)

			backedOff := false
			// We don't know the backoff amount set in the reader, so we try 100 times until it closes.
			for i := 0; i < 100; i++ {
				_, err := red.Read(job)
				if err == reader.ErrBackoffExceeded {
					backedOff = true
					break
				}
			}
			It("should exceed the backoff", func() {
				Expect(backedOff).To(BeTrue())
			})

			Context("sending another job, it should block", func() {
				It("should be gone", func() {
					r, err := red.Read(job)
					Expect(err).To(HaveOccurred())
					Expect(r).To(BeNil())
				})
			})
		})

	})
}

// testReadingReturnsErrorIfNotPingedYet is a helper to test the reader returns an error
// if the caller hasn't called the Ping() method.
func testReadingReturnsErrorIfNotPingedYet(cons Constructor) {
	Context("With a reader initialised", func() {
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

		Context("when creating the object", func() {
			red, err = cons.Object()
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when issuing a read job", func() {
			job := token.New(ctx)
			res, err := red.Read(job)

			It("should error with ErrPingNotCalled", func() {
				Expect(err).To(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(reader.ErrPingNotCalled))
			})
			Specify("result should be empty", func() {
				Expect(res).To(BeNil())
			})
		})
	})
}
