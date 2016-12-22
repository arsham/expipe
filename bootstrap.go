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
// For each routes, we need one engine that has one reader and writes to multiple recorders.
// This is because:
//    1 - Readers should not intercept each other by engaging the recorders.
//    2 - When a reader goes out of scope, we can safely stop the recorders.
// When a recorder goes out of scope, the Engine stops sending to that recorder.
func StartEngines(ctx context.Context, log logrus.FieldLogger, confMap *config.ConfMap) (chan struct{}, error) {
	var wg sync.WaitGroup
	done := make(chan struct{})
	readChanBuff := 1000
	readResChanBuff := 1000
	recChanBuff := 1000
	recResChanBuff := 1000
	for reader, recorders := range confMap.Routes {
		for _, recorder := range recorders {
			wg.Add(1)
			red := confMap.Readers[reader]
			rec := confMap.Recorders[recorder]
			en, err := NewWithConfig(ctx, log, readChanBuff, readResChanBuff, recChanBuff, recResChanBuff, red, rec)
			if err != nil {
				return nil, err
			}
			go func(done <-chan struct{}) {
				<-done
				wg.Done()
			}(en.Start())
		}
	}
	go func() {
		wg.Wait()
		close(done)
	}()
	return done, nil
}
