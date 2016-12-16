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
    "syscall"
    "time"

    "github.com/arsham/expvastic"
    "github.com/arsham/expvastic/lib"
    "github.com/asaskevich/govalidator"
    "github.com/namsral/flag"
    "github.com/olivere/elastic"
)

func main() {
    target, esURL, debugLevel, indexName, typeName, interval, timeout := parseFlags()
    log := lib.GetLogger(*debugLevel)
    bgCtx, cancel := context.WithCancel(context.Background())

    captureSignals(cancel)

    logger := elastic.SetErrorLog(log)
    addr := elastic.SetURL(*esURL)

    client, err := elastic.NewClient(addr, logger)
    if err != nil {
        log.Fatal(err)
    }

    ctx, _ := context.WithTimeout(bgCtx, *timeout)
    _, _, err = client.Ping(*esURL).Do(ctx)
    if err != nil {
        if ctx.Err() != nil {
            log.Fatalf("Timeout: %s - %s", ctx.Err(), err)
        }
        log.Fatalf("Ping failed: %s", err)
    }

    ctx, _ = context.WithTimeout(bgCtx, *timeout)
    conf := expvastic.Conf{
        IndexName:     *indexName,
        TypeName:      *typeName,
        Target:        *target,
        ElasticSearch: *esURL,
        Interval:      *interval,
        Timeout:       *timeout,
        Logger:        log,
    }
    cl, err := expvastic.NewClient(ctx, client, conf)
    if err != nil {
        if ctx.Err() != nil {
            log.Fatalf("Timeout: %s - %s", ctx.Err(), err)
        }
        log.Fatalf("Ping failed: %s", err)
    }

    cl.Start(bgCtx)
}

func parseFlags() (target, esURL, debugLevel, indexName, typeName *string, interval, timeout *time.Duration) {
    var err error
    target = flag.String("target", "localhost:1234/debug/vars", "Target address and port")
    esURL = flag.String("es", "localhost:9200", "Elasticsearch URL and port")
    debugLevel = flag.String("debug", "info", "Debug level")
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
    fmt.Println(*esURL)
    return
}

func captureSignals(cancel context.CancelFunc) {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigCh
        cancel()
        os.Exit(1)
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
