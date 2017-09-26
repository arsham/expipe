// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"strconv"
	"time"

	"github.com/arsham/expipe/reader"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

const (
	name     = "the name"
	typeName = "my type"
)

func testShouldNotChangeTheInput(cons Constructor) {
	Context("With given input", func() {

		endpoint := cons.TestServer().URL
		interval := time.Second
		timeout := time.Second
		backoff := 5
		cons.SetName(name)
		cons.SetTypeName(typeName)
		cons.SetEndpoint(endpoint)
		cons.SetInterval(interval)
		cons.SetTimeout(timeout)
		cons.SetBackoff(backoff)

		red, err := cons.Object()
		It("should not error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
		Specify("name should not be changed", func() {
			Expect(red.Name()).To(Equal(name))
		})
		Specify("type name should not be changed", func() {
			Expect(red.TypeName()).To(Equal(typeName))
		})
		Specify("interval value should not be changed", func() {
			Expect(red.Interval()).To(Equal(interval))
		})
		Specify("timeout value should not be changed", func() {
			Expect(red.Timeout()).To(Equal(timeout))
		})
	})
}

func testNameCheck(cons Constructor) {
	var (
		err      error
		red      reader.DataReader
		name     string
		typeName string
	)

	BeforeEach(func() {
		name = "The name"
		typeName = "the_type_name"
		cons.SetName(name)
		cons.SetTypeName(typeName)
		cons.SetTimeout(time.Hour)
		cons.SetEndpoint(cons.TestServer().URL)
		cons.SetBackoff(5)
	})

	Context("given empty name", func() {
		BeforeEach(func() {
			cons.SetName("")
			red, err = cons.Object()
		})

		It("should error", func() {
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(reader.ErrEmptyName))
		})
		Specify("reader should be nil", func() {
			Expect(red).To(BeNil())
		})
	})
}

func testTypeNameCheck(cons Constructor) {
	var (
		err error
		red reader.DataReader
	)

	cons.SetName(name)
	cons.SetTypeName("")
	cons.SetEndpoint(cons.TestServer().URL)

	Context("given empty type name", func() {
		red, err = cons.Object()

		It("should error", func() {
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(reader.ErrEmptyTypeName))
		})
		Specify("reader should be nil", func() {
			Expect(red).To(BeNil())
		})
	})

}

func testBackoffCheck(cons Constructor) {
	Context("with low backoff value", func() {

		backoff := 3
		cons.SetName("the name")
		cons.SetTypeName("my_type_name")
		cons.SetEndpoint(cons.TestServer().URL)
		cons.SetTimeout(time.Second)
		cons.SetBackoff(backoff)

		red, err := cons.Object()
		Specify("reader to be nil", func() {
			Expect(red).To(BeNil())
		})
		It("should error and mention the backoff value", func() {
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(BeAssignableToTypeOf(reader.ErrLowBackoffValue(0)))
			Expect(err.Error()).To(ContainSubstring(strconv.Itoa(backoff)))
		})
	})
}

func testIntervalCheck(cons Constructor) {
	Context("with zero interval value", func() {
		endpoint := cons.TestServer().URL
		interval := 0
		cons.SetEndpoint(endpoint)
		cons.SetName("the name")
		cons.SetTypeName("my type")
		cons.SetInterval(time.Duration(interval))

		red, err := cons.Object()
		Specify("reader to be nil", func() {
			Expect(red).To(BeNil())
		})
		It("should error and mention the interval value", func() {
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(BeAssignableToTypeOf(reader.ErrLowInterval(0)))
			Expect(err.Error()).To(ContainSubstring(strconv.Itoa(interval)))
		})
	})
}

func testEndpointCheck(cons Constructor) {
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetTimeout(time.Second)
	cons.SetInterval(time.Second)
	cons.SetBackoff(5)

	Context("With invalid endpoint", func() {
		invalidEndpoint := "this is invalid"
		cons.SetEndpoint(invalidEndpoint)

		red, err := cons.Object()
		Specify("reader should be nil", func() {
			Expect(red).To(BeNil())
		})
		It("should error and mention the endpoint", func() {
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(BeAssignableToTypeOf(reader.ErrInvalidEndpoint("")))
			Expect(err.Error()).To(ContainSubstring(invalidEndpoint))
		})
	})

	Context("With empty endpoint", func() {
		cons.SetEndpoint("")
		red, err := cons.Object()
		Specify("reader should be nil", func() {
			Expect(red).To(BeNil())
		})
		It("should error", func() {
			Expect(err).To(HaveOccurred())
			Expect(errors.Cause(err)).To(Equal(reader.ErrEmptyEndpoint))
		})
	})
}
