// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing_test

import (
	"time"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/recorder/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var c *testing.Config
	Describe("NewConfig", func() {
		Context("new object with a set of input", func() {
			var err error
			name := "name"
			log := internal.DiscardLogger()
			endpoint := "http://localhost"
			timeout := time.Second
			backoff := 5
			indexName := "index_name"
			c, err = testing.NewConfig(name, log, endpoint, timeout, backoff, indexName)

			Specify("has all the input in its fields", func() {
				Expect(c.Name()).To(Equal(name))
				Expect(c.Logger()).To(Equal(log))
				Expect(c.Endpoint()).To(Equal(endpoint))
				Expect(c.Timeout()).To(Equal(timeout))
				Expect(c.Backoff()).To(Equal(backoff))
				Expect(c.IndexName()).To(Equal(indexName))
			})
			It("error is nil", func() {
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("NewInstance", func() {
		Context("new recorder set-up from last description", func() {
			r, err := c.NewInstance()
			rec, ok := r.(*testing.Recorder)
			Specify("error should be nil", func() {
				Expect(ok).To(BeTrue())
				Expect(err).NotTo(HaveOccurred())
			})
			Specify("has all the input in its fields", func() {
				Expect(rec.Name()).To(Equal(c.Name()))
				Expect(rec.Endpoint()).To(Equal(c.Endpoint()))
				Expect(rec.Timeout()).To(Equal(c.Timeout()))
				Expect(rec.Backoff()).To(Equal(c.Backoff()))
				Expect(rec.IndexName()).To(Equal(c.IndexName()))
			})
		})
	})
})
