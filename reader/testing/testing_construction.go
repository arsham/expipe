// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"strconv"
	"testing"
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/reader"
	gin "github.com/onsi/ginkgo"
	gom "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

const (
	name     = "the name"
	typeName = "my type"
)

func testShouldNotChangeTheInput(t *testing.T, cons Constructor) {

	endpoint := cons.TestServer().URL
	interval := time.Second
	timeout := time.Second
	backoff := 5
	logger := internal.DiscardLogger()
	cons.SetName(name)
	cons.SetTypeName(typeName)
	cons.SetEndpoint(endpoint)
	cons.SetInterval(interval)
	cons.SetTimeout(timeout)
	cons.SetBackoff(backoff)
	cons.SetLogger(logger)

	red, err := cons.Object()
	if err != nil {
		t.Errorf("want (nil), got (%v)", err)
	}
	if red.Name() != name {
		t.Errorf("want (%s), got (%s)", red.Name(), name)
	}
	if red.TypeName() != typeName {
		t.Errorf("want (%s), got (%s)", red.TypeName(), typeName)
	}
	if red.Interval() != interval {
		t.Errorf("want (%s), got (%s)", red.Interval().String(), interval.String())
	}
	if red.Timeout() != timeout {
		t.Errorf("want (%d), got (%d)", red.Timeout(), timeout)
	}
}

func testNameCheck(cons Constructor) {
	var (
		err      error
		red      reader.DataReader
		name     string
		typeName string
	)

	gin.BeforeEach(func() {
		name = "The name"
		typeName = "the_type_name"
		cons.SetName(name)
		cons.SetTypeName(typeName)
		cons.SetTimeout(time.Hour)
		cons.SetEndpoint(cons.TestServer().URL)
		cons.SetBackoff(5)
	})

	gin.Context("given empty name", func() {
		gin.BeforeEach(func() {
			cons.SetName("")
			red, err = cons.Object()
		})

		gin.It("should error", func() {
			gom.Expect(err).To(gom.HaveOccurred())
			gom.Expect(errors.Cause(err)).To(gom.MatchError(reader.ErrEmptyName))
		})
		gin.Specify("reader should be nil", func() {
			gom.Expect(red).To(gom.BeNil())
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

	gin.Context("given empty type name", func() {
		red, err = cons.Object()

		gin.It("should error", func() {
			gom.Expect(err).To(gom.HaveOccurred())
			gom.Expect(errors.Cause(err)).To(gom.MatchError(reader.ErrEmptyTypeName))
		})
		gin.Specify("reader should be nil", func() {
			gom.Expect(red).To(gom.BeNil())
		})
	})

}

func testBackoffCheck(cons Constructor) {
	gin.Context("with low backoff value", func() {

		backoff := 3
		cons.SetName("the name")
		cons.SetTypeName("my_type_name")
		cons.SetEndpoint(cons.TestServer().URL)
		cons.SetTimeout(time.Second)
		cons.SetBackoff(backoff)

		red, err := cons.Object()
		gin.Specify("reader to be nil", func() {
			gom.Expect(red).To(gom.BeNil())
		})
		gin.It("should error and mention the backoff value", func() {
			gom.Expect(err).To(gom.HaveOccurred())
			gom.Expect(errors.Cause(err)).To(gom.BeAssignableToTypeOf(reader.ErrLowBackoffValue(0)))
			gom.Expect(err.Error()).To(gom.ContainSubstring(strconv.Itoa(backoff)))
		})
	})
}

func testIntervalCheck(cons Constructor) {
	gin.Context("with zero interval value", func() {
		endpoint := cons.TestServer().URL
		interval := 0
		cons.SetEndpoint(endpoint)
		cons.SetName("the name")
		cons.SetTypeName("my type")
		cons.SetInterval(time.Duration(interval))

		red, err := cons.Object()
		gin.Specify("reader to be nil", func() {
			gom.Expect(red).To(gom.BeNil())
		})
		gin.It("should error and mention the interval value", func() {
			gom.Expect(err).To(gom.HaveOccurred())
			gom.Expect(errors.Cause(err)).To(gom.BeAssignableToTypeOf(reader.ErrLowInterval(0)))
			gom.Expect(err.Error()).To(gom.ContainSubstring(strconv.Itoa(interval)))
		})
	})
}

func testEndpointCheck(cons Constructor) {
	cons.SetName("the name")
	cons.SetTypeName("my type")
	cons.SetTimeout(time.Second)
	cons.SetInterval(time.Second)
	cons.SetBackoff(5)

	gin.Context("With invalid endpoint", func() {
		invalidEndpoint := "this is invalid"
		cons.SetEndpoint(invalidEndpoint)

		red, err := cons.Object()
		gin.Specify("reader should be nil", func() {
			gom.Expect(red).To(gom.BeNil())
		})
		gin.It("should error and mention the endpoint", func() {
			gom.Expect(err).To(gom.HaveOccurred())
			gom.Expect(errors.Cause(err)).To(gom.BeAssignableToTypeOf(reader.ErrInvalidEndpoint("")))
			gom.Expect(err.Error()).To(gom.ContainSubstring(invalidEndpoint))
		})
	})

	gin.Context("With empty endpoint", func() {
		cons.SetEndpoint("")
		red, err := cons.Object()
		gin.Specify("reader should be nil", func() {
			gom.Expect(red).To(gom.BeNil())
		})
		gin.It("should error", func() {
			gom.Expect(err).To(gom.HaveOccurred())
			gom.Expect(errors.Cause(err)).To(gom.MatchError(reader.ErrEmptyEndpoint))
		})
	})
}
