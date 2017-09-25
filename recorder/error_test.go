// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package recorder_test

import (
	"strconv"

	"github.com/arsham/expipe/recorder"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("Error Messages", func() {

	Context("With given an ErrInvalidEndpoint", func() {
		msg := "the endpoint"
		e := recorder.ErrInvalidEndpoint(msg)
		It("should contain the message", func() {
			Expect(e.Error()).To(ContainSubstring(msg))
		})
	})

	Context("With given an ErrEndpointNotAvailable", func() {
		endpoint := "the endpoint"
		err := errors.New("my error")
		e := recorder.ErrEndpointNotAvailable{Endpoint: endpoint, Err: err}
		It("should contain the endpoint", func() {
			Expect(e.Error()).To(ContainSubstring(endpoint))
		})
		It("should contain the included error", func() {
			Expect(e.Error()).To(ContainSubstring(err.Error()))
		})
	})

	Context("With given an ErrLowBackoffValue", func() {
		backoff := 5
		e := recorder.ErrLowBackoffValue(backoff)
		It("should contain the backoff value", func() {
			Expect(e.Error()).To(ContainSubstring(strconv.Itoa(backoff)))
		})
	})

	Context("With given an ErrParseTimeOut", func() {
		timeout := "5"
		err := errors.New("my error")
		e := recorder.ErrParseTimeOut{Timeout: timeout, Err: err}
		It("should contain the timeout value", func() {
			Expect(e.Error()).To(ContainSubstring(timeout))
		})

		It("should contain the error message", func() {
			Expect(e.Error()).To(ContainSubstring(err.Error()))
		})
	})

	Context("With given an ErrInvalidIndexName error", func() {

		indexName := "thumb is not an index finger"
		e := recorder.ErrInvalidIndexName(indexName)
		It("should contain the index name", func() {
			Expect(e.Error()).To(ContainSubstring(indexName))
		})
	})

	Context("With given an ErrLowTimeout", func() {
		timeout := 5
		e := recorder.ErrLowTimeout(timeout)
		It("should contain the timeout value", func() {
			Expect(e.Error()).To(ContainSubstring(strconv.Itoa(timeout)))
		})
	})
})
