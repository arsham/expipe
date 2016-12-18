// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arsham/expvastic"
	"github.com/arsham/expvastic/lib"
	"github.com/asaskevich/govalidator"
	"github.com/namsral/flag"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(3)
	target, esURL, debugLevel, indexName, typeName, interval, timeout := parseFlags()
	log := lib.GetLogger(*debugLevel)
	bgCtx, cancel := context.WithCancel(context.Background())
	captureSignals(&wg, cancel)

	ctx, _ := context.WithTimeout(bgCtx, *timeout)
	esClient := getES(ctx, log, *esURL, *indexName)
	reader := getExpvar(log, *target)

	rDone := reader.Start()
	wDone := esClient.Start()
	cl := getEngine(bgCtx, log, reader, esClient, *indexName, *typeName, *interval, *timeout)
	cl.Start()
	<-wDone
	wg.Done()
	<-rDone
	wg.Done()
	wg.Wait()
}

func parseFlags() (target, esURL, debugLevel, indexName, typeName *string, interval, timeout *time.Duration) {
	var err error
	target = flag.String("target", "localhost:1234/debug/vars", "Target address and port")
	esURL = flag.String("es", "localhost:9200", "Elasticsearch URL and port")
	debugLevel = flag.String("loglevel", "info", "Debug level")
	indexName = flag.String("index", "expvastic", "Elasticsearch index name")
	typeName = flag.String("type", "expvastic", "Elasticsearch type name")
	interval = flag.Duration("int", time.Second, "Interval between pulls")
	timeout = flag.Duration("timeout", 30*time.Second, "Elasticsearch communication timeout")
	flag.Parse()
	if *esURL, err = validateURL(*esURL); err != nil {
		fmt.Println("Invalid ElasticSearch URL")
		flag.Usage()
		os.Exit(1)
	}

	if *target, err = validateURL(*target); err != nil {
		fmt.Println("Invalid target URL")
		flag.Usage()
		os.Exit(1)
	}
	return
}

func captureSignals(wg *sync.WaitGroup, cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
		wg.Done()
	}()

}

func validateURL(url string) (string, error) {
	if govalidator.IsURL(url) {
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "http://" + url
		}
		return url, nil
	}
	return "", fmt.Errorf("Invalid url: %s", url)

}

func getES(ctx context.Context, log logrus.FieldLogger, esURL, indexName string) *expvastic.ElasticSearch {
	esClient, err := expvastic.NewElasticSearch(ctx, log, esURL, indexName)
	if err != nil {
		if ctx.Err() != nil {
			log.Fatalf("Timeout: %s - %s", ctx.Err(), err)
		}
		log.Fatalf("Ping failed: %s", err)
	}
	return esClient
}

func getExpvar(log logrus.FieldLogger, target string) *expvastic.ExpvarReader {
	r, err := expvastic.NewExpvarReader(log, expvastic.NewCtxReader(target))
	if err != nil {
		log.Fatalf("Error creating the reader: %s", err)
	}
	return r
}

func getEngine(
	bgCtx context.Context,
	log logrus.FieldLogger,
	reader *expvastic.ExpvarReader,
	esClient *expvastic.ElasticSearch,
	indexName,
	typeName string,
	interval,
	timeout time.Duration,
) *expvastic.Engine {
	conf := expvastic.Conf{
		TargetReader: reader,
		Recorder:     esClient,
		IndexName:    indexName,
		TypeName:     typeName,
		Interval:     interval,
		Timeout:      timeout,
		Logger:       log,
	}
	return expvastic.NewEngine(bgCtx, conf)
}
