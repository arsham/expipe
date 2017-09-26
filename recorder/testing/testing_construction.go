// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"strconv"
	"time"

	"github.com/arsham/expipe/recorder"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

func testShouldNotChangeTheInput(cons Constructor) {
	Context("With given input", func() {
		name := "recorder name"
		indexName := "recorder_index_name"
		endpoint := cons.TestServer().URL
		timeout := time.Second
		backoff := 5
		cons.SetName(name)
		cons.SetIndexName(indexName)
		cons.SetEndpoint(endpoint)
		cons.SetTimeout(timeout)
		cons.SetBackoff(backoff)

		rec, err := cons.Object()
		It("should not error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
		Specify("name should not be changed", func() {
			Expect(rec.Name()).To(Equal(name))
		})
		Specify("index name should not be changed", func() {
			Expect(rec.IndexName()).To(Equal(indexName))
		})
		Specify("timeout should not be changed", func() {
			Expect(rec.Timeout()).To(Equal(timeout))
		})
	})
}

func testBackoffCheck(cons Constructor) {
	Context("with low backoff value", func() {
		backoff := 3
		cons.SetName("the name")
		cons.SetIndexName("my_index_name")
		cons.SetEndpoint(cons.TestServer().URL)
		cons.SetTimeout(time.Second)

		cons.SetBackoff(backoff)
		rec, err := cons.Object()

		Specify("recorder to be nil", func() {
			Expect(rec).To(BeNil())
		})
		It("should error and mention the backoff value", func() {
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(BeAssignableToTypeOf(recorder.ErrLowBackoffValue(0)))
			Expect(err.Error()).To(ContainSubstring(strconv.Itoa(backoff)))
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

	BeforeEach(func() {
		name = "The name"
		indexName = "the_index_name"
		cons.SetName(name)
		cons.SetIndexName(indexName)
		cons.SetTimeout(time.Hour)
		cons.SetEndpoint(cons.TestServer().URL)
		cons.SetBackoff(5)
	})

	Context("given empty name", func() {
		BeforeEach(func() {
			cons.SetName("")
			rec, err = cons.Object()
		})

		It("should error", func() {
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(recorder.ErrEmptyName))
		})
		Specify("recorder should be nil", func() {
			Expect(rec).To(BeNil())
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

	BeforeEach(func() {
		name = "the name"
		indexName = "index_name"
		cons.SetName(name)
		cons.SetIndexName(indexName)
		cons.SetTimeout(time.Hour)
		cons.SetEndpoint(cons.TestServer().URL)
		cons.SetBackoff(5)
	})

	Context("given invalid index name", func() {
		BeforeEach(func() {
			cons.SetIndexName("aa bb")
			rec, err = cons.Object()
		})

		It("should error", func() {
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(BeAssignableToTypeOf(recorder.ErrInvalidIndexName("")))
		})
		Specify("recorder should be nil", func() {
			Expect(rec).To(BeNil())
		})
	})
}

func testEndpointCheck(cons Constructor) {
	Context("With invalid endpoint", func() {

		invalidEndpoint := "this is invalid"
		cons.SetName("the name")
		cons.SetIndexName("my_index_name")
		cons.SetTimeout(time.Hour)
		cons.SetBackoff(5)
		cons.SetEndpoint(invalidEndpoint)
		rec, err := cons.Object()

		Specify("recorder should be nil", func() {
			Expect(rec).To(BeNil())
		})
		It("should error and mention the endpoint", func() {
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(BeAssignableToTypeOf(recorder.ErrInvalidEndpoint("")))
			Expect(err.Error()).To(ContainSubstring(invalidEndpoint))
		})
	})
}
