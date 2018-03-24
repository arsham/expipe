// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"time"

	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/token"
	gin "github.com/onsi/ginkgo"
	gom "github.com/onsi/gomega"
)

// testReaderReceivesJob is a test helper to test the reader can receive jobs
func testReaderReceivesJob(cons Constructor) {
	gin.Context("Having a reader setup", func() {
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

		gin.Context("when creating the reader", func() {
			red, err = cons.Object()
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
		})

		gin.Context("when pinging", func() {
			err = red.Ping()
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
		})

		gin.Context("When reading from the endpoint", func() {

			ctx := context.Background()
			result, err := red.Read(token.New(ctx))
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
			gin.Specify("result should not be nil", func() {
				gom.Expect(result).NotTo(gom.BeNil())
			})
			gin.Specify("result.ID should not be empty", func() {
				gom.Expect(result.ID).NotTo(gom.BeEmpty())
			})
			gin.Specify("result.TypeName should not be empty", func() {
				gom.Expect(result.TypeName).NotTo(gom.BeEmpty())
			})
			gin.Specify("result.Content should not be nil", func() {
				gom.Expect(result.Content).NotTo(gom.BeNil())
			})
			gin.Specify("result.Mapper should not be nil", func() {
				gom.Expect(result.Mapper).NotTo(gom.BeNil())
			})
		})
	})
}

// testReaderReturnsSameID is a test helper to test the reader returns the same ID in the response
func testReaderReturnsSameID(cons Constructor) {
	gin.Context("Having a reader set up", func() {
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

		gin.Context("when creating the reader", func() {
			red, err = cons.Object()
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
		})
		gin.Context("when pinging the endpoint", func() {
			err = red.Ping()
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
		})

		gin.Context("when reading from the endpoint", func() {

			ctx := context.Background()
			job := token.New(ctx)
			result, err := red.Read(job)
			gin.It("should not error", func() {
				gom.Expect(err).NotTo(gom.HaveOccurred())
			})
			gin.Specify("result should not be nil", func() {
				gom.Expect(result).NotTo(gom.BeNil())
			})
			gin.Specify("result.ID should be job.ID", func() {
				gom.Expect(result.ID).To(gom.Equal(job.ID()))
			})
		})
	})
}
