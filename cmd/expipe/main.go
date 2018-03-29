// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package main

import (
	"context"
	"os"
	"time"

	"github.com/arsham/expipe"
	"github.com/arsham/expipe/cmd/expipe/app"
	flags "github.com/jessevdk/go-flags"
)

func main() {
	var done chan struct{}
	flags.Parse(&app.Opts)
	log, confSlice, err := app.Config()
	if err != nil {
		log.Fatalf(err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	app.CaptureSignals(cancel, sigCh, os.Exit, 10*time.Second)
	done, err = expipe.StartEngines(ctx, log, confSlice)
	if err != nil {
		log.Fatalf(err.Error())
	}
	<-done
}
