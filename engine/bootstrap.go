// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine

import (
	"context"
	"sync"

	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/tools"
	"github.com/arsham/expipe/tools/config"
	"github.com/pkg/errors"
)

// StartEngines creates some Engines and returns a channel that closes it when
// it's done its work. For each routes, we need one engine that has multiple
// readers and writes to one recorder. When all recorders of one reader
// go out of scope, the Engine stops that reader because there is no destination.
// Each Engine is ran in its own goroutine.
func StartEngines(ctx context.Context, log tools.FieldLogger, conf *config.ConfMap) (chan struct{}, error) {
	// TODO: return a slice of error
	var (
		wg       sync.WaitGroup
		leastOne bool
		err      error
	)
	done := make(chan struct{})
	if conf == nil {
		return nil, errors.New("confMap cannot be nil")
	}
	for recorder, readers := range conf.Routes {
		var en *Engine
		en, err = getEngine(ctx, log, conf, recorder, readers)
		if err != nil {
			log.Warn(err)
			continue
		}
		wg.Add(1)
		leastOne = true
		go func(en *Engine) {
			Start(en)
			log.Infof("Engine's work (%s) has finished", en)
			wg.Done()
		}(en)
	}
	if !leastOne {
		return nil, err
	}
	go func() {
		wg.Wait()
		close(done)
	}()
	return done, err
}

func getEngine(ctx context.Context, log tools.FieldLogger, conf *config.ConfMap, recorder string, readers []string) (*Engine, error) {
	rec := conf.Recorders[recorder]
	if rec == nil {
		return nil, errors.New("empty recorder")
	}
	reds := make([]reader.DataReader, 0)
	for _, reader := range readers {
		if _, ok := conf.Readers[reader]; !ok {
			continue
		}
		red := conf.Readers[reader]
		reds = append(reds, red)
	}
	if len(reds) == 0 { // TEST:
		return nil, ErrNoReader
	}
	return New(
		WithCtx(ctx),
		WithRecorder(rec),
		WithReaders(reds...),
		WithLogger(log),
	)
}
