// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expipe_test

import (
	"fmt"

	"github.com/arsham/expipe"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error messages", func() {
	Context("with ErrPing", func() {

		Context("given one pair", func() {
			name := "divine"
			body := "is a myth"
			err := expipe.ErrPing{name: fmt.Errorf(body)}
			Specify("error message should contain the body and the name of the error", func() {
				Expect(err.Error()).To(ContainSubstring(name))
				Expect(err.Error()).To(ContainSubstring(body))
			})
		})

		Context("given two pairs", func() {
			name1 := "divine"
			body1 := "is a myth"
			name2 := "science"
			body2 := "just works!"
			err := expipe.ErrPing{
				name1: fmt.Errorf(body1),
				name2: fmt.Errorf(body2),
			}
			Specify("error message should contain the body and the name of the error pairs", func() {
				Expect(err.Error()).To(ContainSubstring(name1))
				Expect(err.Error()).To(ContainSubstring(body1))
				Expect(err.Error()).To(ContainSubstring(name2))
				Expect(err.Error()).To(ContainSubstring(body2))
			})
		})
	})
})
