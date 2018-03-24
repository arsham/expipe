// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expipe

import (
	"context"
	"sync"

	"github.com/arsham/expipe/config"
	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/reader"
	"github.com/pkg/errors"
)

// StartEngines creates some Engines and returns a channel that closes it when
// it's done its work. For each routes, we need one engine that has multiple
// readers and writes to one recorder. When all recorders of one reader
// go out of scope, the Engine stops that reader because there is no destination.
// Each Engine is ran in its own goroutine.
func StartEngines(ctx context.Context, log internal.FieldLogger, confMap *config.ConfMap) (chan struct{}, error) {
	// TODO: return a slice of error
	var (
		wg       sync.WaitGroup
		leastOne bool
		err      error
	)
	done := make(chan struct{})
	if confMap == nil {
		return nil, errors.New("confMap cannot be nil")
	}
	for recorder, readers := range confMap.Routes {
		var en *Engine
		en, err = getEngine(ctx, log, confMap, recorder, readers)
		if err != nil {
			log.Warn(err)
			continue
		}
		wg.Add(1)
		leastOne = true
		go func(en *Engine) {
			en.Start()
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

func getEngine(ctx context.Context, log internal.FieldLogger, confMap *config.ConfMap, recorder string, readers []string) (*Engine, error) {
	recMap := confMap.Recorders[recorder]
	if recMap == nil {
		return nil, errors.New("empty recorder")
	}
	rec, err := recMap.NewInstance()
	if err != nil {
		return nil, errors.Wrap(err, "bootstrapping recorder")
	}
	reds := make([]reader.DataReader, 1)
	for _, reader := range readers {
		if _, ok := confMap.Readers[reader]; !ok {
			continue
		}
		red, err := confMap.Readers[reader].NewInstance()
		if err != nil {
			return nil, errors.Wrap(err, "new engine with config")
		}
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
