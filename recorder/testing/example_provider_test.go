// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"github.com/arsham/expipe/recorder/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExampleProvider", func() {
	Describe("GetRecorder", func() {
		Context("with a sane url", func() {
			url := "http://localhost"
			r := testing.GetRecorder(url)
			It("returns a recorder", func() {
				Expect(r).To(BeAssignableToTypeOf(&testing.Recorder{}))
			})
			It("is not nil", func() {
				Expect(r).NotTo(BeNil())
			})
			It("has name and index set", func() {
				Expect(r.Name()).NotTo(BeEmpty())
				Expect(r.IndexName()).NotTo(BeEmpty())
			})
			It("has a logger", func() {
				Expect(r.Logger()).NotTo(BeNil())
			})
			It("has timeout > 0", func() {
				Expect(r.Timeout()).To(BeNumerically(">", 0))
			})
			It("has backoff >= 5", func() {
				Expect(r.Backoff()).To(BeNumerically(">=", 5))
			})
		})
		Context("with a bad url", func() {
			url := "bad url"
			It("panics", func() {
				Expect(func() { testing.GetRecorder(url) }).To(Panic())
			})
		})
	})
})
