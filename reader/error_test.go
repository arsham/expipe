// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader_test

import (
	"strconv"

	"github.com/arsham/expipe/reader"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("Error Messages", func() {

	Context("With given an ErrInvalidEndpoint", func() {
		msg := "the endpoint"
		e := reader.ErrInvalidEndpoint(msg)
		It("should contain the error message", func() {
			Expect(e.Error()).To(ContainSubstring(msg))
		})
	})

	Context("With given an ErrEndpointNotAvailable", func() {
		endpoint := "the endpoint"
		err := errors.New("my error")
		e := reader.ErrEndpointNotAvailable{Endpoint: endpoint, Err: err}
		It("should contain the endpoint", func() {
			Expect(e.Error()).To(ContainSubstring(endpoint))
		})
		It("should contain the included error", func() {
			Expect(e.Error()).To(ContainSubstring(err.Error()))
		})
	})

	Context("With given an ErrLowBackoffValue", func() {
		backoff := 5
		e := reader.ErrLowBackoffValue(backoff)
		It("should contain the backoff value", func() {
			Expect(e.Error()).To(ContainSubstring(strconv.Itoa(backoff)))
		})
	})

	Context("With given an ErrLowInterval", func() {
		interval := 5
		e := reader.ErrLowInterval(interval)
		It("should contain the interval value", func() {
			Expect(e.Error()).To(ContainSubstring(strconv.Itoa(interval)))
		})
	})

	Context("With given an ErrLowTimeout", func() {
		timeout := 5
		e := reader.ErrLowTimeout(timeout)
		It("should contain the timeout value", func() {
			Expect(e.Error()).To(ContainSubstring(strconv.Itoa(timeout)))
		})
	})
})
