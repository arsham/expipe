// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/config"
	"github.com/arsham/expvastic/lib"
	"github.com/namsral/flag"
	"github.com/spf13/viper"
)

var (
	confFile = flag.String("c", "expvastic", "Confuration file. Should be in yaml format without the extension")
)

func main() {
	log := lib.GetLogger("info")
	flag.Parse()
	v := viper.New()
	v.SetConfigName(*confFile)
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	err := v.ReadInConfig()
	if err != nil {
		log.Fatalf("Config file not found or contains error: %s. Falling back to env/flags", err)
	}

	confSlice, err := config.LoadYAML(log, v)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	captureSignals(cancel)
	done, err := expvastic.StartEngines(ctx, log, confSlice)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func captureSignals(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

}
