// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/arsham/expipe/internal/datatype"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Container", func() {
	Describe("New", func() {
		Context("when creating a new Container out of a list", func() {
			l := []datatype.DataType{datatype.FloatType{}}
			c := datatype.New(l)
			It("contains the list", func() {
				Expect(c.List()).To(Equal(l))
			})
		})
	})

	Describe(".List", func() {
		Context("having a list in the container", func() {
			l := []datatype.DataType{datatype.FloatType{}}
			c := datatype.New(l)
			It("returns the same list", func() {
				Expect(c.List()).To(Equal(l))
			})
		})
	})

	Describe(".Len", func() {
		Context("having a list in the container", func() {
			l := []datatype.DataType{datatype.FloatType{}, datatype.FloatType{}}
			c := datatype.New(l)
			It("returns length of the list", func() {
				Expect(c.Len()).To(Equal(len(l)))
			})
		})
	})

	Describe(".Add", func() {
		Context("having an empty list", func() {
			c := datatype.New(nil)
			firstElm := datatype.FloatType{}
			Specify("adding a new element makes the list to grow by one", func() {
				l := c.Len()
				c.Add(firstElm)
				Expect(c.Len()).To(Equal(l + 1))
			})
			Specify("adding multiple elements grows the list to the amount of elements", func() {
				l := c.Len()
				c.Add(datatype.FloatType{}, datatype.FloatType{}, datatype.FloatType{})
				Expect(c.Len()).To(Equal(l + 3))
			})
			Specify("adding the same element does not makes the list to grow", func() {
				l := c.Len()
				c.Add(firstElm)
				Expect(c.Len()).To(Equal(l + 1))
			})
		})
	})

	Describe(".Error", func() {
		Context("having a Container with an error embodied", func() {
			err := errors.New("an error")
			c := datatype.Container{Err: err}
			It("returns the same error", func() {
				Expect(c.Error()).To(Equal(err))
			})
			Specify("returning again will return the same error", func() {
				Expect(c.Error()).To(Equal(err))
			})
		})
	})

	Describe(".Bytes", func() {
		now := time.Now()
		Context("having an empty container", func() {
			c := datatype.Container{}
			It("returns only the timestamp", func() {
				expected := fmt.Sprintf(`{"@timestamp":"%s"}`, now.Format(datatype.TimeStampFormat))
				Expect(c.Bytes(now)).To(Equal([]byte(expected)))
			})
		})

		Context("having a container with one element", func() {
			elm := datatype.FloatType{Key: "key", Value: 66.6}
			c := datatype.New(nil)
			c.Add(elm)

			It("should contain the provided timestamp", func() {
				Expect(c.Bytes(now)).To(ContainSubstring(now.Format(datatype.TimeStampFormat)))
			})
			It("should contain the element's byte string rep", func() {
				Expect(c.Bytes(now)).To(ContainSubstring(string(elm.Bytes())))
			})
		})
	})
})
