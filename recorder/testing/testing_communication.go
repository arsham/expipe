// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package testing

import (
	"context"
	"time"

	"github.com/arsham/expipe/internal/datatype"
	"github.com/arsham/expipe/internal/token"
	"github.com/arsham/expipe/recorder"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// testRecorderReceivesPayload tests the recorder receives the payload correctly.
func testRecorderReceivesPayload(cons Constructor) {
	Context("With correctly created recorder", func() {
		var (
			err error
			rec recorder.DataRecorder
		)
		ctx := context.Background()
		cons.SetName("the name")
		cons.SetIndexName("my_index")
		cons.SetTimeout(time.Second)
		cons.SetBackoff(5)
		cons.SetEndpoint(cons.TestServer().URL)

		Context("by creating recorder", func() {
			rec, err = cons.Object()
			It("should not error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			Specify("recorder should not be nil", func() {
				Expect(rec).NotTo(BeNil())
			})
		})

		Context("when pinging the endpoint", func() {
			err = rec.Ping()
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when sending payload", func() {
			p := datatype.New([]datatype.DataType{})
			payload := &recorder.Job{
				ID:        token.NewUID(),
				Payload:   p,
				IndexName: "my index",
				TypeName:  "my type",
				Time:      time.Now(),
			}

			received := make(chan struct{})
			go func() {
				rec.Record(ctx, payload)
				received <- struct{}{}
			}()
			It("should receive the payload", func(done Done) {
				<-received
				close(done)
			}, 5)
		})
	})
}

// testRecorderSendsResult tests the recorder send the results to the endpoint.
func testRecorderSendsResult(cons Constructor) {
	Context("with initially created recorder", func() {
		var (
			err error
			rec recorder.DataRecorder
		)
		ctx := context.Background()
		cons.SetName("the name")
		cons.SetIndexName("index_name")
		cons.SetTimeout(time.Second)
		cons.SetBackoff(15)
		cons.SetEndpoint(cons.TestServer().URL)

		Context("when getting the object", func() {
			rec, err = cons.Object()
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("when pinging the endpoint", func() {
			err = rec.Ping()
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when payload is received and result is recorded", func() {
			p := datatype.New([]datatype.DataType{&datatype.StringType{Key: "test", Value: "test"}})
			payload := &recorder.Job{
				ID:        token.NewUID(),
				Payload:   p,
				IndexName: "my_index",
				TypeName:  "my_type",
				Time:      time.Now(),
			}

			err = rec.Record(ctx, payload)
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
}
