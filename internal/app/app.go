// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/engine"
	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/reader/expvar"
	"github.com/arsham/expipe/recorder"
	"github.com/arsham/expipe/recorder/elasticsearch"
	"github.com/arsham/expipe/tools"
	"github.com/arsham/expipe/tools/config"
	flags "github.com/jessevdk/go-flags"
	"github.com/spf13/viper"
)

// TODO: change the log to FieldLogger

var (
	log *tools.Logger
)

// Opts is the command line flag struct.
// IDEA: create an interactive wizard for creating a config file.
var Opts struct {
	ConfFile  string        `short:"c" long:"config" env:"CONFIG" default:"" description:"Configuration file. Should be in yaml format without the extension."`
	Reader    string        `long:"reader" env:"READER" default:"localhost:1234/debug/vars" description:"Target address and port"`
	Recorder  string        `long:"recorder" env:"RECORDER" default:"localhost:9200" description:"Elasticsearch URL and port"`
	LogLevel  string        `long:"loglevel" env:"LOGLEVEL" default:"info" description:"Log level"`
	IndexName string        `long:"index" env:"INDEX" default:"expipe" description:"Elasticsearch index name"`
	TypeName  string        `long:"type" env:"TYPE" default:"expipe" description:"Elasticsearch type name"`
	Interval  time.Duration `long:"int" env:"INT" default:"1s" description:"Interval between pulls from the target"`
	Timeout   time.Duration `long:"timeout" env:"TIMEOUT" default:"30s" description:"Communication time-outs to both reader and recorder"`
}

// Main is the entrypoint of the application. It is been called from main.main.
// It captures SIGINT or SIGTERM signals to terminate the app.
func Main() {
	_, conf, err := Config()
	if err != nil {
		log.Fatalf(err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	CaptureSignals(cancel, sigCh, os.Exit, 1*time.Second)
	Bootstrap(ctx, log, conf)
}

// Config returns the ConfMap from a file if it was set in the command flags.
func Config() (*tools.Logger, *config.ConfMap, error) {
	flags.Parse(&Opts)
	log = tools.GetLogger("info")
	if Opts.ConfFile == "" {
		log = tools.GetLogger(Opts.LogLevel)
		conf, err := fromFlags()
		return log, conf, err
	}
	conf, err := fromConfig(Opts.ConfFile)
	return log, conf, err
}

// Bootstrap sets up an instance of the Service and starts it. It waits until
// the Service signals its work has been finished.
func Bootstrap(ctx context.Context, log tools.FieldLogger, conf *config.ConfMap) {
	s := engine.Service{
		Ctx:  ctx,
		Log:  log,
		Conf: conf,
	}

	done, err := s.Start()
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	<-done
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
		Readers:   make(map[string]reader.DataReader, 1),
		Recorders: make(map[string]recorder.DataRecorder, 1),
	}

	confMap.Recorders["elasticsearch"], err = elasticsearch.New(
		recorder.WithLogger(log),
		recorder.WithName("recorder"),
		recorder.WithEndpoint(Opts.Recorder),
		recorder.WithTimeout(Opts.Timeout),
		recorder.WithIndexName(Opts.IndexName),
	)
	if err != nil {
		return nil, err
	}
	confMap.Readers["expvar"], err = expvar.New(
		reader.WithLogger(log),
		reader.WithName("expvar"),
		reader.WithTypeName(Opts.TypeName),
		reader.WithEndpoint(Opts.Reader),
		reader.WithInterval(Opts.Interval),
		reader.WithTimeout(Opts.Timeout),
		reader.WithMapper(datatype.DefaultMapper()),
	)
	if err != nil {
		return nil, err
	}
	confMap.Routes = make(map[string][]string)
	confMap.Routes["expvar"] = make([]string, 1)
	confMap.Routes["expvar"][0] = "elasticsearch"
	return confMap, nil
}

// CaptureSignals cancels the context if receives the SIGINT or SIGTERM signal
// through sigCh, and exits with calling exit(130).
func CaptureSignals(cancel context.CancelFunc, sigCh chan os.Signal, exit func(int), timeout time.Duration) {
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		go func() {
			<-time.After(timeout)
			exit(130)
		}()
		cancel()
	}()
}
