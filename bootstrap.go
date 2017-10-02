// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expipe

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/internal/config"
	"github.com/arsham/expipe/reader"
	"github.com/pkg/errors"
)

// StartEngines creates some Engines and returns a channel that closes it when it's done its work.
// For each routes, we need one engine that has multiple readers and writes to one recorder.
// When all recorders of one reader go out of scope, the Engine stops that reader because there
// is no destination.
func StartEngines(ctx context.Context, log internal.FieldLogger, confMap *config.ConfMap) (chan struct{}, error) {
	var (
		wg       sync.WaitGroup
		leastOne uint32
		err      error
	)
	done := make(chan struct{})

	if confMap == nil {
		return nil, errors.New("confMap cannot be nil")
	}

	for recorder, readers := range confMap.Routes {
		var en *Engine
		recMap := confMap.Recorders[recorder]
		if recMap == nil {
			return nil, errors.New("empty recorder")
		}
		rec, errR := recMap.NewInstance()
		if errR != nil {
			return nil, errors.Wrap(errR, "bootstrapping recorder")
		}

		reds := make([]reader.DataReader, 1)
		for _, reader := range readers {
			if _, ok := confMap.Readers[reader]; !ok {
				continue
			}
			red, errR := confMap.Readers[reader].NewInstance()
			if errR != nil {
				return nil, errors.Wrap(errR, "new engine with config")
			}
			reds = append(reds, red)
		}
		if len(reds) == 0 {
			return nil, ErrNoReader
		}

		en, err = New(
			SetCtx(ctx),
			SetRecorder(rec),
			SetReaders(reds...),
			SetLogger(log),
		)
		if err != nil {
			log.Warn(err)
			continue
		}
		wg.Add(1)
		atomic.StoreUint32(&leastOne, uint32(1))
		go func() {
			en.Start()
			log.Infof("Engine's work (%s) has finished", en)
			wg.Done()
		}()
	}
	if atomic.LoadUint32(&leastOne) < 1 {
		return nil, err
	}
	go func() {
		wg.Wait()
		close(done)
	}()
	return done, nil
}
