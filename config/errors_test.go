// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config_test

import (
	"fmt"

	"github.com/arsham/expipe/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

const (
	nilStr = "<nil>"
)

var _ = Describe("Error messages", func() {
	Context("with StructureErr", func() {
		Describe("given a nil StructureErr", func() {
			err := (*config.StructureErr)(nil)
			It("should print the 'nilStr' value", func() {
				Expect(err.Error()).To(Equal(nilStr))
			})
		})

		Describe("given a section, reason and body", func() {
			section := "this section"
			reason := "the reason"
			body := "whatever body is there"
			DescribeTable("should contain section, reason and body", func(err error) {
				Expect(err.Error()).To(ContainSubstring(section))
				Expect(err.Error()).To(ContainSubstring(reason))
				Expect(err.Error()).To(ContainSubstring(body))
			},
				Entry("StructureErr",
					&config.StructureErr{
						Section: section,
						Reason:  reason,
						Err:     fmt.Errorf(body),
					},
				),
				Entry("ErrNotSpecified",
					&config.ErrNotSpecified{
						Section: section,
						Reason:  reason,
						Err:     fmt.Errorf(body),
					},
				),
			)
		})
	})

	Context("With ErrNotSupported", func() {
		Describe("having a string as the value", func() {
			body := "god"
			err := config.ErrNotSupported(body)
			It("should print the content in error message", func() {
				Expect(err.Error()).To(ContainSubstring(body))
			})
		})
	})
})
