// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"strconv"
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/recorder"
	gin "github.com/onsi/ginkgo"
	gom "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

func testShouldNotChangeTheInput(cons Constructor) {
	gin.Context("With given input", func() {
		name := "recorder name"
		indexName := "recorder_index_name"
		endpoint := cons.TestServer().URL
		timeout := time.Second
		backoff := 5
		logger := internal.DiscardLogger()
		cons.SetName(name)
		cons.SetIndexName(indexName)
		cons.SetEndpoint(endpoint)
		cons.SetTimeout(timeout)
		cons.SetBackoff(backoff)
		cons.SetLogger(logger)

		rec, err := cons.Object()
		gin.It("should not error", func() {
			gom.Expect(err).NotTo(gom.HaveOccurred())
		})
		gin.Specify("name should not be changed", func() {
			gom.Expect(rec.Name()).To(gom.Equal(name))
		})
		gin.Specify("index name should not be changed", func() {
			gom.Expect(rec.IndexName()).To(gom.Equal(indexName))
		})
		gin.Specify("timeout should not be changed", func() {
			gom.Expect(rec.Timeout()).To(gom.Equal(timeout))
		})
		gin.Specify("logger should not be changed", func() {
			gom.Expect(logger).To(gom.BeIdenticalTo(logger))
		})

	})
}

func testBackoffCheck(cons Constructor) {
	gin.Context("with low backoff value", func() {
		backoff := 3
		cons.SetName("the name")
		cons.SetIndexName("my_index_name")
		cons.SetEndpoint(cons.TestServer().URL)
		cons.SetTimeout(time.Second)

		cons.SetBackoff(backoff)
		rec, err := cons.Object()

		gin.Specify("recorder to be nil", func() {
			gom.Expect(rec).To(gom.BeNil())
		})
		gin.It("should error and mention the backoff value", func() {
			gom.Expect(err).To(gom.HaveOccurred())
			gom.Expect(errors.Cause(err)).To(gom.BeAssignableToTypeOf(recorder.ErrLowBackoffValue(0)))
			gom.Expect(err.Error()).To(gom.ContainSubstring(strconv.Itoa(backoff)))
		})
	})
}

func testNameCheck(cons Constructor) {
	var (
		err       error
		rec       recorder.DataRecorder
		name      string
		indexName string
	)

	gin.BeforeEach(func() {
		name = "The name"
		indexName = "the_index_name"
		cons.SetName(name)
		cons.SetIndexName(indexName)
		cons.SetTimeout(time.Hour)
		cons.SetEndpoint(cons.TestServer().URL)
		cons.SetBackoff(5)
	})

	gin.Context("given empty name", func() {
		gin.BeforeEach(func() {
			cons.SetName("")
			rec, err = cons.Object()
		})

		gin.It("should error", func() {
			gom.Expect(err).To(gom.HaveOccurred())
			gom.Expect(errors.Cause(err)).To(gom.MatchError(recorder.ErrEmptyName))
		})
		gin.Specify("recorder should be nil", func() {
			gom.Expect(rec).To(gom.BeNil())
		})
	})
}

func testIndexNameCheck(cons Constructor) {
	var (
		err       error
		rec       recorder.DataRecorder
		name      string
		indexName string
	)

	gin.BeforeEach(func() {
		name = "the name"
		indexName = "index_name"
		cons.SetName(name)
		cons.SetIndexName(indexName)
		cons.SetTimeout(time.Hour)
		cons.SetEndpoint(cons.TestServer().URL)
		cons.SetBackoff(5)
	})

	gin.Context("given invalid index name", func() {
		gin.BeforeEach(func() {
			cons.SetIndexName("aa bb")
			rec, err = cons.Object()
		})

		gin.It("should error", func() {
			gom.Expect(err).To(gom.HaveOccurred())
			gom.Expect(errors.Cause(err)).To(gom.BeAssignableToTypeOf(recorder.ErrInvalidIndexName("")))
		})
		gin.Specify("recorder should be nil", func() {
			gom.Expect(rec).To(gom.BeNil())
		})
	})
}

func testEndpointCheck(cons Constructor) {
	gin.Context("With invalid endpoint", func() {

		invalidEndpoint := "this is invalid"
		cons.SetName("the name")
		cons.SetIndexName("my_index_name")
		cons.SetTimeout(time.Hour)
		cons.SetBackoff(5)
		cons.SetEndpoint(invalidEndpoint)
		rec, err := cons.Object()

		gin.Specify("recorder should be nil", func() {
			gom.Expect(rec).To(gom.BeNil())
		})
		gin.It("should error and mention the endpoint", func() {
			gom.Expect(err).To(gom.HaveOccurred())
			gom.Expect(errors.Cause(err)).To(gom.BeAssignableToTypeOf(recorder.ErrInvalidEndpoint("")))
			gom.Expect(err.Error()).To(gom.ContainSubstring(invalidEndpoint))
		})
	})
}
