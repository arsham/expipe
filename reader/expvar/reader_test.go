// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvar

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
)

func TestExpvarReaderErrors(t *testing.T) {
	log := lib.DiscardLogger()
	ctx, cancel := context.WithCancel(context.Background())
	ctxReader := reader.NewMockCtxReader("nowhere")
	ctxReader.ContextReadFunc = func(ctx context.Context) (*http.Response, error) {
		return nil, fmt.Errorf("Error")
	}
	jobChan := make(chan context.Context)
	errorChan := make(chan communication.ErrorMessage)
	resultChan := make(chan *reader.ReadJobResult)

	mapper := &datatype.MapConvertMock{}
	red, _ := NewExpvarReader(log, ctxReader, mapper, jobChan, resultChan, errorChan, "my_reader", "example_type", time.Second, time.Second)
	red.Start(ctx)
	defer cancel()

	red.JobChan() <- communication.NewReadJob(ctx)
	select {
	case res := <-red.ResultChan():
		if res.Res != nil {
			t.Errorf("expecting no results, got(%v)", res.Res)
		}
	case <-time.After(100 * time.Millisecond):
	}

	select {
	case err := <-errorChan:
		if err.Error() == "" {
			t.Error("expecting error, got nothing")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expecting an error result back, got nothing")
	}
	cancel()
}

func TestExpvarReaderClosesStream(t *testing.T) {
	log := lib.DiscardLogger()
	ctxReader := reader.NewMockCtxReader("nowhere")
	ctx, cancel := context.WithCancel(context.Background())
	jobChan := make(chan context.Context)
	resultChan := make(chan *reader.ReadJobResult)
	errorChan := make(chan communication.ErrorMessage)

	mapper := &datatype.MapConvertMock{}
	red, _ := NewExpvarReader(log, ctxReader, mapper, jobChan, resultChan, errorChan, "my_reader", "example_type", time.Second, time.Second)
	done := red.Start(ctx)
	jobChan <- communication.NewReadJob(ctx)
	<-errorChan
	cancel()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("The channel was not closed in time")
	}
}
