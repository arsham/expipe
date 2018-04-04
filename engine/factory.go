// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package engine

import (
	"context"
	"sync"

	"github.com/arsham/expipe/recorder"
	"github.com/arsham/expipe/tools"
	"github.com/arsham/expipe/tools/config"
	"github.com/pkg/errors"
)

// Service initialises Engines.
// Configure injects the input values into the Operator by calling each function
// on it.
type Service struct {
	Log       tools.FieldLogger
	Ctx       context.Context
	Conf      *config.ConfMap
	Configure func(...func(Engine) error) (Engine, error)
}

// Start creates some Engines and returns a channel that closes it when
// it's done its work. For each routes, we need one engine that has multiple
// readers and writes to one recorder. When all recorders of one reader
// go out of scope, the Engine stops that reader because there is no destination.
// Each Engine is ran in its own goroutine.
func (s *Service) Start() (chan struct{}, error) {
	// TODO: return a slice of error
	var (
		wg       sync.WaitGroup
		leastOne bool
		err      error
	)
	if s.Configure == nil {
		s.Configure = New
	}
	done := make(chan struct{})
	if s.Conf == nil {
		return nil, errors.New("confMap cannot be nil")
	}
	for reader, recorders := range s.Conf.Routes {
		var en Engine

		en, err = s.engine(reader, recorders)
		if err != nil {
			s.Log.Warn(err)
			continue
		}
		wg.Add(1)
		leastOne = true
		go func(en Engine) {
			done := Start(en)
			<-done
			s.Log.Infof("Engine's work (%s) has finished", en)
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

func (s *Service) engine(reader string, recorders []string) (Engine, error) {
	red := s.Conf.Readers[reader]
	if red == nil {
		return nil, errors.New("empty reader")
	}
	recs := make([]recorder.DataRecorder, 0)
	for _, rec := range recorders {
		if r, ok := s.Conf.Recorders[rec]; ok {
			recs = append(recs, r)
		}
	}
	if len(recs) == 0 {
		return nil, ErrNoRecorder
	}
	return s.Configure(
		WithCtx(s.Ctx),
		WithReader(red),
		WithRecorders(recs...),
		WithLogger(s.Log),
	)
}
