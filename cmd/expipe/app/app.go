// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/arsham/expipe/config"
	"github.com/arsham/expipe/internal"
	"github.com/arsham/expipe/reader/expvar"
	"github.com/arsham/expipe/recorder/elasticsearch"
	"github.com/spf13/viper"
)

// TODO: change the log to FieldLogger

var (
	log *internal.Logger
)

// Opts is the command line flag struct.
// IDEA: create an interactive wizard for creating a config file.
var Opts struct {
	ConfFile  string        `long:"c" env:"CONFIG" default:"" description:"Configuration file. Should be in yaml format without the extension."`
	Reader    string        `long:"reader" env:"READER" default:"localhost:1234/debug/vars" description:"Target address and port"`
	Recorder  string        `long:"recorder" env:"RECORDER" default:"localhost:9200" description:"Elasticsearch URL and port"`
	LogLevel  string        `long:"loglevel" env:"LOGLEVEL" default:"info" description:"Log level"`
	IndexName string        `long:"index" env:"INDEX" default:"expipe" description:"Elasticsearch index name"`
	TypeName  string        `long:"type" env:"TYPE" default:"expipe" description:"Elasticsearch type name"`
	Interval  time.Duration `long:"int" env:"INT" default:"1s" description:"Interval between pulls from the target"`
	Timeout   time.Duration `long:"timeout" env:"TIMEOUT" default:"30s" description:"Communication time-outs to both reader and recorder"`
	Backoff   int           `long:"backoff" env:"BACKOFF" default:"15" description:"After this amount, it will give up accessing unresponsive endpoints"`
}

// Config returns the ConfMap from a file if it was set in the command flags.
func Config() (*internal.Logger, *config.ConfMap, error) {
	log = internal.GetLogger("info")
	if Opts.ConfFile == "" {
		log = internal.GetLogger(Opts.LogLevel)
		conf, err := fromFlags()
		return log, conf, err
	}
	conf, err := fromConfig(Opts.ConfFile)
	return log, conf, err
}

// setting up from config file
func fromConfig(confFile string) (*config.ConfMap, error) {
	v := viper.New()
	v.SetConfigName(confFile)
	v.SetConfigType("yaml") // PLAN: Also read from toml, json etcd, consul, etc.
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

// setting up from command flags
func fromFlags() (*config.ConfMap, error) {
	var err error
	confMap := &config.ConfMap{
		Readers:   make(map[string]config.ReaderConf, 1),
		Recorders: make(map[string]config.RecorderConf, 1),
	}

	confMap.Recorders["elasticsearch"], err = elasticsearch.NewConfig(log, "elasticsearch", Opts.Recorder, Opts.Timeout, Opts.Backoff, Opts.IndexName)
	if err != nil {
		return nil, err
	}
	r := strings.SplitN(Opts.Reader, "/", 2)
	if len(r) != 2 {
		return nil, fmt.Errorf("reader endpoint should have a route: %s", Opts.Reader)
	}
	confMap.Readers["expvar"], err = expvar.NewConfig(log, "expvar", Opts.TypeName, r[0], r[1], Opts.Interval, Opts.Timeout, Opts.Backoff, "")
	if err != nil {
		return nil, err
	}
	confMap.Routes = make(map[string][]string)
	confMap.Routes["expvar"] = make([]string, 1)
	confMap.Routes["expvar"][0] = "elasticsearch"
	return confMap, nil
}

// CaptureSignals cancels the context if receives the SIGINT or SIGTERM.
func CaptureSignals(cancel context.CancelFunc) {
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
