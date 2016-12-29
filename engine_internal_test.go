// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
	"bytes"
	"context"
	"io/ioutil"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/test"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/recorder"
)

type errMsg string

func (e errMsg) Error() string { return string(e) }

// inspectLogs checks if the niddle is found in the entries
// the entries might have been stacked, we need to iterate over.
func inspectLogs(entries []*logrus.Entry, niddle string) (all string, found bool) {
	var res []string
	for _, field := range entries {
		if strings.Contains(field.Message, niddle) {
			return "", true
		}
		res = append(res, field.Message)
	}
	return strings.Join(res, ", "), false
}

func withReaders(ctx context.Context, log logrus.FieldLogger, cancel context.CancelFunc) (engine *Engine, jobChan chan context.Context, redJobResChan chan *reader.ReadJobResult, errorChan chan communication.ErrorMessage) {
	payloadChan := make(chan *recorder.RecordJob)
	errorChan = make(chan communication.ErrorMessage)
	rec, _ := recorder.NewSimpleRecorder(ctx, log, payloadChan, errorChan, "recorder_test", "nowhere", "indexName", time.Hour)
	jobChan = make(chan context.Context)
	redJobResChan = make(chan *reader.ReadJobResult)
	engine = &Engine{
		name:          "test_engine",
		ctx:           ctx,
		log:           log,
		recorder:      rec,
		readerResChan: redJobResChan,
		errorChan:     errorChan,
	}
	return
}

func TestEventLoopCatchesReaderError(t *testing.T) {
	log, hook := test.NewNullLogger()
	log.Level = logrus.ErrorLevel

	ctx, cancel := context.WithCancel(context.Background())
	e, jobChan, redJobResChan, errorChan := withReaders(ctx, log, cancel)
	red, err := reader.NewSimpleReader(lib.DiscardLogger(), "http://127.0.0.1:9200", jobChan, redJobResChan, errorChan, "reader_name", "typeName", time.Hour, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}

	redStop := make(communication.StopChannel)
	e.setReaders(map[reader.DataReader]communication.StopChannel{red: redStop})

	jobID := communication.NewJobID()
	errMsg := errMsg("an error happened")
	recorded := make(chan struct{})

	// Testing the engine catches errors
	red.StartFunc = func(stop communication.StopChannel) {
		go func() {
			<-jobChan
			errorChan <- communication.ErrorMessage{ID: jobID, Err: errMsg}
			recorded <- struct{}{} // important, otherwise the test might not be valid
		}()
		go func() {
			s := <-stop
			s <- struct{}{}
		}()
	}

	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()
	red.JobChan() <- ctx

	select {
	case <-recorded:
	case <-time.After(5 * time.Second):
		t.Error("expected to record, didn't happen")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}

	if _, found := inspectLogs(hook.Entries, errMsg.Error()); !found {
		// sometimes it takes time for logrus to register the error, trying again
		time.Sleep(500 * time.Millisecond)
		if all, found := inspectLogs(hook.Entries, errMsg.Error()); !found {
			t.Errorf("want (%s) in the error, got (%v)", errMsg.Error(), all)
		}
	}
}

func TestEventLoopOneReaderSendsPayload(t *testing.T) {
	log := lib.DiscardLogger()

	ctx, cancel := context.WithCancel(context.Background())
	e, jobChan, redJobResChan, errorChan := withReaders(ctx, log, cancel)
	red, err := reader.NewSimpleReader(lib.DiscardLogger(), "http://127.0.0.1:9200", jobChan, redJobResChan, errorChan, "reader_name", "typeName", time.Hour, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}
	redStop := make(communication.StopChannel)
	e.setReaders(map[reader.DataReader]communication.StopChannel{red: redStop})

	jobID := communication.NewJobID()
	recorded := make(chan struct{})

	// testing engine send the payload to the recorder
	red.StartFunc = func(stop communication.StopChannel) {
		go func() {
			<-red.JobChan()
			resp := &reader.ReadJobResult{
				ID:       jobID,
				Res:      ioutil.NopCloser(bytes.NewBuffer([]byte(`{"devil":666}`))),
				TypeName: red.TypeName(),
				Mapper:   red.Mapper(),
			}
			red.ResultChan() <- resp
			go func() {
				s := <-stop
				s <- struct{}{}
			}()

		}()
	}

	rec := e.recorder.(*recorder.SimpleRecorder)
	rec.StartFunc = func(stop communication.StopChannel) {
		go func() {
			recordedPayload := <-rec.PayloadChan()
			if recordedPayload.ID != jobID {
				t.Errorf("want (%s), got (%s)", jobID, recordedPayload.ID)
			}
			recorded <- struct{}{}
		}()
		go func() {
			s := <-stop
			s <- struct{}{}
		}()
	}

	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()

	jobChan <- ctx
	select {
	case <-recorded:
	case <-time.After(5 * time.Second):
		t.Error("expected to record, didn't happen")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}
}

func TestEventLoopRecorderGoesOutOfScope(t *testing.T) {
	log, hook := test.NewNullLogger()
	log.Level = logrus.DebugLevel

	ctx, cancel := context.WithCancel(context.Background())
	e, jobChan, redJobResChan, errorChan := withReaders(ctx, log, cancel)
	red1, _ := reader.NewSimpleReader(lib.DiscardLogger(), "http://127.0.0.1:9200", jobChan, redJobResChan, errorChan, "reader_name", "typeName", time.Hour, time.Hour)
	red2, _ := reader.NewSimpleReader(lib.DiscardLogger(), "http://127.0.0.1:9200", jobChan, redJobResChan, errorChan, "reader2_name", "typeName", time.Hour, time.Hour)

	e.setReaders(map[reader.DataReader]communication.StopChannel{red1: make(communication.StopChannel), red2: make(communication.StopChannel)})
	red1.StartFunc = func(stop communication.StopChannel) {
		go func() {
			s := <-stop
			s <- struct{}{}
		}()
	}

	red2.StartFunc = func(stop communication.StopChannel) {
		go func() {
			s := <-stop
			s <- struct{}{}
		}()
	}

	rec := e.recorder.(*recorder.SimpleRecorder)
	rec.StartFunc = func(stop communication.StopChannel) {
		go func() {
			s := <-stop
			s <- struct{}{}
		}()
	}

	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}

	if _, found := inspectLogs(hook.Entries, recorderGone); !found {
		// sometimes it takes time for logrus to register the error, trying again
		time.Sleep(500 * time.Millisecond)
		if all, found := inspectLogs(hook.Entries, recorderGone); !found {
			t.Errorf("want (%s) in the error, got (%v)", recorderGone, all)
		}
	}

	select {
	case jobChan <- communication.NewReadJob(ctx):
		t.Error("expected the engine to close the readers")
	case <-time.After(20 * time.Millisecond):
	}
}

func TestEventLoopClosingContext(t *testing.T) {
	log, hook := test.NewNullLogger()
	log.Level = logrus.DebugLevel

	ctx, cancel := context.WithCancel(context.Background())
	e, jobChan, redJobResChan, errorChan := withReaders(ctx, log, cancel)
	red, err := reader.NewSimpleReader(lib.DiscardLogger(), "http://127.0.0.1:9200", jobChan, redJobResChan, errorChan, "reader_name", "typeName", time.Hour, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}
	stop := make(communication.StopChannel)
	e.setReaders(map[reader.DataReader]communication.StopChannel{red: stop})

	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("expected the engine to quit gracefully")
	}

	if _, found := inspectLogs(hook.Entries, contextCanceled); !found {
		// sometimes it takes time for logrus to register the error, trying again
		time.Sleep(500 * time.Millisecond)
		if all, found := inspectLogs(hook.Entries, contextCanceled); !found {
			t.Errorf("want (%s) in the error, got (%v)", contextCanceled, all)
		}
	}
}

func TestEventLoopMultipleReadersSendPayload(t *testing.T) {
	log := lib.DiscardLogger()
	log.Level = logrus.DebugLevel

	ctx, cancel := context.WithCancel(context.Background())
	e, jobChan, redJobResChan, errorChan := withReaders(ctx, log, cancel)
	red1, _ := reader.NewSimpleReader(lib.DiscardLogger(), "http://127.0.0.1:9200", jobChan, redJobResChan, errorChan, "reader1_name", "typeName", time.Hour, time.Hour)
	red2, _ := reader.NewSimpleReader(lib.DiscardLogger(), "http://127.0.0.1:9200", jobChan, redJobResChan, errorChan, "reader2_name", "typeName", time.Hour, time.Hour)
	red1Stop := make(communication.StopChannel)
	red2Stop := make(communication.StopChannel)
	e.setReaders(map[reader.DataReader]communication.StopChannel{red1: red1Stop, red2: red2Stop})

	jobID1 := communication.NewJobID()
	jobID2 := communication.NewJobID()
	recorded := make(chan struct{})

	// testing engine send the payloads to the recorder
	red1.StartFunc = func(stop communication.StopChannel) {
		go func() {
			<-red1.JobChan()
			resp := &reader.ReadJobResult{
				ID:       jobID1,
				Res:      ioutil.NopCloser(bytes.NewBuffer([]byte(`{"devil":666}`))),
				TypeName: red1.TypeName(),
				Mapper:   red1.Mapper(),
			}
			red1.ResultChan() <- resp
			go func() {
				s := <-stop
				s <- struct{}{}
			}()
		}()
	}

	red2.StartFunc = func(stop communication.StopChannel) {
		go func() {
			<-red2.JobChan()
			resp := &reader.ReadJobResult{
				ID:       jobID2,
				Res:      ioutil.NopCloser(bytes.NewBuffer([]byte(`{"beelzebub":666}`))),
				TypeName: red2.TypeName(),
				Mapper:   red2.Mapper(),
			}
			red2.ResultChan() <- resp
			go func() {
				s := <-stop
				s <- struct{}{}
			}()
		}()
	}

	rec := e.recorder.(*recorder.SimpleRecorder)
	rec.StartFunc = func(stop communication.StopChannel) {
		go func() {
			recordedPayload := <-rec.PayloadChan()
			if recordedPayload.ID != jobID1 && recordedPayload.ID != jobID2 {
				t.Errorf("want one of (%s, %s), got (%s)", jobID1, jobID2, recordedPayload.ID)
			}
			if recordedPayload.ID != jobID1 && recordedPayload.ID != jobID2 {
				t.Errorf("want one of (%s, %s), got (%s)", jobID1, jobID2, recordedPayload.ID)
			}
			recorded <- struct{}{}
			recorded <- struct{}{}
		}()
		go func() {
			s := <-stop
			s <- struct{}{}
		}()
	}
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		e.Start()
		wg.Done()
	}()

	jobChan <- ctx
	select {
	case <-recorded:
	case <-time.After(5 * time.Second):
		t.Error("expected to record, didn't happen")
	}
	select {
	case <-recorded:
	case <-time.After(5 * time.Second):
		t.Error("expected to record, didn't happen")
	}

	cancel()
	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}

}

func TestStartReadersTicking(t *testing.T) {
	log := lib.DiscardLogger()

	ctx, cancel := context.WithCancel(context.Background())
	e, jobChan, redJobResChan, errorChan := withReaders(ctx, log, cancel)
	red, err := reader.NewSimpleReader(lib.DiscardLogger(), "http://127.0.0.1:9200", jobChan, redJobResChan, errorChan, "reader_name", "typeName", 10*time.Millisecond, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error occurred during reader creation: %v", err)
	}
	redStop := make(communication.StopChannel)
	e.setReaders(map[reader.DataReader]communication.StopChannel{red: redStop})

	recorded := make(chan struct{})

	// Testing the engine ticks and sends a job request to the reader
	// There is no need for the actual job
	red.StartFunc = func(stop communication.StopChannel) {
		go func() {
			<-jobChan
			jobID := communication.NewJobID()
			errorChan <- communication.ErrorMessage{ID: jobID, Err: errMsg("blah blah")}
			recorded <- struct{}{} // important, otherwise the test might not be valid
		}()
		go func() {
			s := <-stop
			s <- struct{}{}
		}()
	}

	done := make(chan struct{})
	go func() {
		e.Start()
		done <- struct{}{}
	}()

	select {
	case <-recorded:
	case <-time.After(2 * time.Second):
		t.Error("expected to record, didn't happen")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("expected the engine to quit gracefully")
	}
}
