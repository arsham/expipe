// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
	"context"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic/config"
)

// StartEngines creates some Engines and returns a channel that closes it when it's done its work.
// For each routes, we need one engine that has multiple readers and writes to one recorder.
// When all recorders of one reader go out of scope, the Engine stops that reader because there
// is no destination.
func StartEngines(ctx context.Context, log logrus.FieldLogger, confMap *config.ConfMap) (chan struct{}, error) {
	var wg sync.WaitGroup
	done := make(chan struct{})

	for recorder, readers := range confMap.Routes {
		for _, reader := range readers {
			red := confMap.Readers[reader]
			rec := confMap.Recorders[recorder]
			en, err := NewWithConfig(ctx, log, rec, red)
			if err != nil {
				return nil, err
			}
			go func() {
				wg.Add(1)
				en.Start()
				log.Infof("Engine %s has finished", en.name)
				wg.Done()
			}()
		}
	}
	go func() {
		wg.Wait()
		close(done)
	}()
	return done, nil
}
