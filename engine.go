// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
	"context"
	"expvar"
	"fmt"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/config"
	"github.com/arsham/expvastic/reader"
	"github.com/arsham/expvastic/recorder"
)

var (
	numGoroutines   = expvar.NewInt("Number Of Goroutines")
	expRecorders    = expvar.NewInt("Recorders")
	expReaders      = expvar.NewInt("Readers")
	readJobs        = expvar.NewInt("Read Jobs")
	recordJobs      = expvar.NewInt("Record Jobs")
	erroredJobs     = expvar.NewInt("Error Jobs")
	recorderGone    = "recorder is gone now"
	contextCanceled = "context has been cancelled"
)

// Engine represents an engine that receives information from readers and ships them to a recorder.
// The Engine is allowed to change the index and type names at will.
// When the context times out or cancelled, the engine will close the the job channels by calling its
// stop method. It will send a stop signal to readers and recorders asking them to finish their jobs.
// It will timeout the stop signals if it doesn't receive a response.
// Note that we could create a channel and distribute the recorders payload, but we didn't because there
// is no way to find out which recorder errors right after the payload has been sent.
// IMPORTANT: the readers should not close their streams, the Engine closes them.
type Engine struct {
	log           logrus.FieldLogger
	ctx           context.Context                   // Will call stop() when this context is cancelled/timed-out. This is a new context from the parent.
	name          string                            // Name identifier for this engine.
	shutdown      sync.Once                         // The Engine signals itself to shut-down.
	recorder      recorder.DataRecorder             // Records to ElasticSearch client.
	errorChan     <-chan communication.ErrorMessage // Recorder and all Readers will send their errors through this channel
	readerResChan <-chan *reader.ReadJobResult      // Readers report their results through this channel.

	redmu   sync.RWMutex
	readers map[reader.DataReader]communication.StopChannel // Map of active readers to their stop signals
}

// NewWithConfig instantiates reader and recorders from the configurations and sends them
// to the NewWithReadRecorder. The engine's work starts from there.
// readChanBuff, readResChanBuff, recChanBuff are the channel buffer amount.
// Please refer to the benchmarks how to choose the best values.
func NewWithConfig(ctx context.Context, log logrus.FieldLogger,
	readChanBuff, readResChanBuff, recChanBuff int,
	recorderConf config.RecorderConf, readers ...config.ReaderConf) (*Engine, error) {

	reds := make([]reader.DataReader, len(readers))
	errorChan := make(chan communication.ErrorMessage, recChanBuff+(len(readers)*readChanBuff)) // large enough so both reader and red can report
	readerResChan := make(chan *reader.ReadJobResult, len(readers)*readResChanBuff)

	for i, redConf := range readers {
		jobChan := make(chan context.Context, readChanBuff)
		red, err := redConf.NewInstance(ctx, jobChan, readerResChan, errorChan)
		if err != nil {
			return nil, err
		}
		reds[i] = red
	}

	recorderPayloadChan := make(chan *recorder.RecordJob, recChanBuff)
	rec, err := recorderConf.NewInstance(ctx, recorderPayloadChan, errorChan)
	if err != nil {
		return nil, err
	}
	return NewWithReadRecorder(ctx, log, errorChan, readerResChan, rec, reds...)
}

// NewWithReadRecorder creates an instance an Engine with already made reader and recorders.
// It streams all readers payloads to the recorder.
// Returns an error if there are recorders with the same name, or any of them have no name.
func NewWithReadRecorder(ctx context.Context, log logrus.FieldLogger, errorChan <-chan communication.ErrorMessage,
	readerResChan <-chan *reader.ReadJobResult, rec recorder.DataRecorder, reds ...reader.DataReader) (*Engine, error) {

	readerNames := make([]string, len(reds))
	readerMap := make(map[reader.DataReader]communication.StopChannel, len(reds))
	seenNames := make(map[string]struct{}, len(reds))

	for i, red := range reds {
		if _, ok := seenNames[red.Name()]; ok {
			return nil, ErrDuplicateRecorderName
		}

		seenNames[red.Name()] = struct{}{}
		readerNames[i] = red.Name()
		readerMap[red] = make(communication.StopChannel)
	}

	// just to be cute
	engineName := fmt.Sprintf("( %s >-x-<< %s )", rec.Name(), strings.Join(readerNames, ","))
	log = log.WithField("engine", engineName)
	cl := &Engine{
		name:          engineName,
		ctx:           ctx,
		errorChan:     errorChan,
		recorder:      rec,
		readers:       readerMap,
		readerResChan: readerResChan,
		log:           log,
	}
	log.Debug("started the engine")
	return cl, nil
}

func (e *Engine) setReaders(readers map[reader.DataReader]communication.StopChannel) {
	e.redmu.Lock()
	defer e.redmu.Unlock()
	e.readers = readers
}
