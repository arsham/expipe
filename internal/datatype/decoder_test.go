// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype_test

import (
	"github.com/pkg/errors"

	"github.com/arsham/expipe/internal/datatype"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("JobResultDataTypes", func() {

	Context("given a mapper", func() {
		mapper := datatype.DefaultMapper()

		Context("when byte slice is not a valid JSON object", func() {

			DescribeTable("it should error", func(input []byte) {
				c := datatype.JobResultDataTypes(input, mapper)
				Expect(c.Error()).To(HaveOccurred())
			},
				Entry("missing leading {", []byte(`"memstats": {"PauseNs":[666,777]}}`)),
				Entry("missing ending }", []byte(`{"memstats": {"PauseNs":[666,777]}`)),
				Entry("simple string", []byte(`"memstats PauseNs 666 777"`)),
			)
		})

		By("reading the values")

		Context("when input's value type doesn't match the mapper", func() {

			DescribeTable("should error with ErrUnidentifiedJason", func(input []byte) {
				c := datatype.JobResultDataTypes(input, mapper)
				Expect(errors.Cause(c.Error())).To(Equal(datatype.ErrUnidentifiedJason))
			},
				Entry("string instead of float", []byte(`{"memstats": {"PauseNs":["666"]}}`)),
				Entry("float instead of int", []byte(`{"memstats": {"TotalAlloc":[666.5]}}`)),
			)
		})

		Context("when data is matched", func() {

			DescribeTable("should match the input without any errors", func(input []byte, exp ...datatype.DataType) {
				c := datatype.JobResultDataTypes(input, mapper)
				l := datatype.New(nil)
				l.Add(exp...)
				Expect(errors.Cause(c.Error())).NotTo(HaveOccurred())
				Expect(c.List()).To(ConsistOf(l.List()))
			},
				Entry("one value", []byte(`{"memstats": {"TotalAlloc":666}}`), &datatype.MegaByteType{Key: "memstats.TotalAlloc", Value: 666}),
				Entry("multiple values", []byte(`{"memstats": {"TotalAlloc":666, "HeapIdle":777}}`), &datatype.MegaByteType{Key: "memstats.TotalAlloc", Value: 666}, &datatype.MegaByteType{Key: "memstats.HeapIdle", Value: 777}),
			)
		})
	})
})
