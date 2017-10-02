// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/arsham/expipe"
	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/internal/config"
	"github.com/arsham/expipe/reader/expvar"
	"github.com/arsham/expipe/recorder/elasticsearch"
	"github.com/namsral/flag"
	"github.com/spf13/viper"
)

var (
	log               *internal.Logger
	shallStartEngines = true // for testing purposes
	confFile          = flag.String("c", "", "Configuration file. Should be in yaml format without the extension.")
	reader            = flag.String("reader", "localhost:1234/debug/vars", "Target address and port")
	recorder          = flag.String("recorder", "localhost:9200", "Elasticsearch URL and port")
	debugLevel        = flag.String("loglevel", "info", "Log level")
	indexName         = flag.String("index", "expipe", "Elasticsearch index name")
	typeName          = flag.String("app", "expipe", "App name, which will be the Elasticsearch type name")
	interval          = flag.Duration("int", time.Second, "Interval between pulls from the target")
	timeout           = flag.Duration("timeout", 30*time.Second, "Communication time-outs to both reader and recorder")
	backoff           = flag.Int("backoff", 15, "After this amount, it will give up accessing unresponsive endpoints")
	// ExitCommand is used for replacing during tests.
	ExitCommand = func(msg string) {
		log.Fatalf(msg)
	}
)

func main() {
	var (
		confSlice *config.ConfMap
		err       error
		done      chan struct{}
	)
	flag.Parse()

	if *confFile == "" {
		log = internal.GetLogger(*debugLevel)
		confSlice, err = fromFlags()
	} else {
		log = internal.GetLogger("info")
		confSlice, err = fromConfig(*confFile)
	}

	if err != nil {
		ExitCommand(err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if shallStartEngines {
		captureSignals(cancel)
		done, err = expipe.StartEngines(ctx, log, confSlice)
		if err != nil {
			ExitCommand(err.Error())
			return
		}
	}

	if shallStartEngines {
		<-done
	}
}

func fromConfig(confFile string) (*config.ConfMap, error) {
	v := viper.New()
	v.SetConfigName(confFile)
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("reading config file: %s", err)
	}

	confSlice, err := config.LoadYAML(log, v)
	if err != nil {
		return nil, err
	}
	return confSlice, nil
}

func fromFlags() (*config.ConfMap, error) {
	var err error
	confMap := &config.ConfMap{
		Readers:   make(map[string]config.ReaderConf, 1),
		Recorders: make(map[string]config.RecorderConf, 1),
	}

	confMap.Recorders["elasticsearch"], err = elasticsearch.NewConfig(log, "elasticsearch", *recorder, *timeout, *backoff, *indexName)
	if err != nil {
		return nil, err
	}
	r := strings.SplitN(*reader, "/", 2)
	if len(r) != 2 {
		return nil, fmt.Errorf("reader endpoint should have a route: %s", *reader)
	}
	confMap.Readers["expvar"], err = expvar.NewConfig(log, "expvar", *typeName, r[0], r[1], *interval, *timeout, *backoff, "")
	if err != nil {
		return nil, err
	}
	confMap.Routes = make(map[string][]string)
	confMap.Routes["expvar"] = make([]string, 1)
	confMap.Routes["expvar"][0] = "elasticsearch"
	return confMap, nil
}

func captureSignals(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		go func() {
			<-time.After(10 * time.Second)
			os.Exit(1)
		}()
		cancel()
	}()
}
