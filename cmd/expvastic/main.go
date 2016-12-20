// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/config"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/reader/expvar"
	"github.com/arsham/expvastic/recorder/elasticsearch"
	"github.com/namsral/flag"
	"github.com/spf13/viper"
)

var (
	confFile = flag.String("c", "", "Confuration file. Should be in yaml format without the extension.")

	reader     = flag.String("reader", "localhost:1234/debug/vars", "Target address and port")
	recorder   = flag.String("recorder", "localhost:9200", "Elasticsearch URL and port")
	debugLevel = flag.String("loglevel", "info", "Log level")
	indexName  = flag.String("index", "expvastic", "Elasticsearch index name")
	typeName   = flag.String("app", "expvastic", "App name, which will be the Elasticsearch type name")
	interval   = flag.Duration("int", time.Second, "Interval between pulls from the target")
	timeout    = flag.Duration("timeout", 30*time.Second, "Communication timeouts to both reader and recorder")
	backoff    = flag.Int("backoff", 15, "After this amount, it will give up accessing unresponsive endpoints") // TODO: implement!
)

func main() {
	var log *logrus.Logger
	var confSlice *config.ConfMap
	flag.Parse()
	if *confFile == "" {
		log = lib.GetLogger(*debugLevel)
		confSlice = fromFlags(log)
	} else {
		log = lib.GetLogger("info")
		confSlice = fromConfig(log, *confFile)
	}

	ctx, cancel := context.WithCancel(context.Background())
	captureSignals(cancel)
	done, err := expvastic.StartEngines(ctx, log, confSlice)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func fromConfig(log *logrus.Logger, confFile string) *config.ConfMap {
	v := viper.New()
	v.SetConfigName(confFile)
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	err := v.ReadInConfig()
	if err != nil {
		log.Fatalf("Config file not found or contains error: %s", err)
	}

	confSlice, err := config.LoadYAML(log, v)
	if err != nil {
		log.Fatal(err)
	}
	return confSlice
}

func fromFlags(log *logrus.Logger) *config.ConfMap {
	var err error
	confMap := &config.ConfMap{
		Readers:   make(map[string]config.ReaderConf, 1),
		Recorders: make(map[string]config.RecorderConf, 1),
	}

	confMap.Recorders["elasticsearch"], err = elasticsearch.NewConfig("elasticsearch", log, *recorder, *interval, *timeout, *backoff, *indexName, *typeName)
	if err != nil {
		log.Fatal(err)
	}
	r := strings.SplitN(*reader, "/", 2)
	confMap.Readers["expvar"], err = expvar.NewConfig("expvar", log, r[0], r[1], *interval, *timeout, *backoff)
	if err != nil {
		log.Fatal(err)
	}
	confMap.Routes = make(map[string][]string)
	confMap.Routes["expvar"] = make([]string, 1)
	confMap.Routes["expvar"][0] = "elasticsearch"
	return confMap
}

func captureSignals(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()
}
