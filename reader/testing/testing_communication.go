// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"time"

	"github.com/arsham/expipe/internal/token"
	"github.com/arsham/expipe/reader"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// testReaderReceivesJob is a test helper to test the reader can receive jobs
func testReaderReceivesJob(cons Constructor) {
	Context("Having a reader setup", func() {
		var (
			err error
			red reader.DataReader
		)
		cons.SetName("the name")
		cons.SetTypeName("my type")
		cons.SetEndpoint(cons.TestServer().URL)
		cons.SetInterval(time.Hour)
		cons.SetTimeout(time.Hour)
		cons.SetBackoff(5)

		Context("when creating the reader", func() {
			red, err = cons.Object()
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when pinging", func() {
			err = red.Ping()
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When reading from the endpoint", func() {

			ctx := context.Background()
			result, err := red.Read(token.New(ctx))
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			Specify("result should not be nil", func() {
				Expect(result).NotTo(BeNil())
			})
			Specify("result.ID should not be empty", func() {
				Expect(result.ID).NotTo(BeEmpty())
			})
			Specify("result.TypeName should not be empty", func() {
				Expect(result.TypeName).NotTo(BeEmpty())
			})
			Specify("result.Content should not be nil", func() {
				Expect(result.Content).NotTo(BeNil())
			})
			Specify("result.Mapper should not be nil", func() {
				Expect(result.Mapper).NotTo(BeNil())
			})
		})
	})
}

// testReaderReturnsSameID is a test helper to test the reader returns the same ID in the response
func testReaderReturnsSameID(cons Constructor) {
	Context("Having a reader set up", func() {
		var (
			red reader.DataReader
			err error
		)
		cons.SetName("the name")
		cons.SetTypeName("my type")
		cons.SetEndpoint(cons.TestServer().URL)
		cons.SetInterval(time.Hour)
		cons.SetTimeout(time.Hour)
		cons.SetBackoff(5)

		Context("when creating the reader", func() {
			red, err = cons.Object()
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("when pinging the endpoint", func() {
			err = red.Ping()
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when reading from the endpoint", func() {

			ctx := context.Background()
			job := token.New(ctx)
			result, err := red.Read(job)
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			Specify("result should not be nil", func() {
				Expect(result).NotTo(BeNil())
			})
			Specify("result.ID should be job.ID", func() {
				Expect(result.ID).To(Equal(job.ID()))
			})
		})
	})
}
