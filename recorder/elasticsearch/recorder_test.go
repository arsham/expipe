// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package elasticsearch_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/arsham/expvastic/communication"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/recorder"
	"github.com/arsham/expvastic/recorder/elasticsearch"
	"github.com/olivere/elastic"
)

func isTravis() bool {
	return os.Getenv("TRAVIS") != ""
}

func destroyIndex(t *testing.T, endpoint, indexName string) {
	log := lib.DiscardLogger()
	addr := elastic.SetURL(endpoint)
	logger := elastic.SetErrorLog(log)
	timeout := time.Second

	client, err := elastic.NewClient(
		addr,
		logger,
		elastic.SetHealthcheckTimeoutStartup(timeout),
		elastic.SetSnifferTimeout(timeout),
		elastic.SetHealthcheckTimeout(timeout),
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, _, err = client.Ping(endpoint).Do(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.DeleteIndex(indexName).Do(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func setupWithURL(t *testing.T, URL string, indexName string, errorChan chan communication.ErrorMessage) (ctx context.Context, rec *elasticsearch.Recorder) {
	var err error
	log := lib.DiscardLogger()
	ctx = context.Background()
	payloadChan := make(chan *recorder.RecordJob)

	timeout := 10 * time.Millisecond
	if isTravis() {
		timeout = 10 * time.Second
	}

	rec, err = elasticsearch.NewRecorder(ctx, log, payloadChan, errorChan, "reader_example", URL, indexName, timeout)
	if err != nil {
		t.Fatal(err)
	}
	return ctx, rec
}

func setup(t *testing.T, errorChan chan communication.ErrorMessage) (ctx context.Context, rec *elasticsearch.Recorder, teardown func()) {
	endpoint := "http://127.0.0.1:9200"
	indexName := "expvastic_test"
	ctx, rec = setupWithURL(t, endpoint, indexName, errorChan)
	return ctx, rec, func() {
		destroyIndex(t, endpoint, indexName)
	}
}

func TestElasticsearchRecorder(t *testing.T) {
	recorder.TestRecorderEssentials(t, func(testCase int) (context.Context, recorder.DataRecorder, error, chan communication.ErrorMessage, func()) {
		switch testCase {
		case recorder.RecorderReceivesPayloadTestCase:
			errorChan := make(chan communication.ErrorMessage)
			ctx, rec, teardown := setup(t, make(chan communication.ErrorMessage))
			return ctx, rec, nil, errorChan, teardown

		case recorder.RecorderSendsResultTestCase:
			errorChan := make(chan communication.ErrorMessage)
			ctx, rec, teardown := setup(t, errorChan)
			return ctx, rec, nil, errorChan, teardown

		case recorder.RecorderClosesTestCase:
			errorChan := make(chan communication.ErrorMessage)
			ctx, rec, teardown := setup(t, errorChan)
			return ctx, rec, nil, errorChan, teardown

		case recorder.RecorderErrorsOnUnavailableEndpointTestCase:
			var err error
			log := lib.DiscardLogger()
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

			timeout := 1 * time.Microsecond
			if isTravis() {
				timeout = 10 * time.Second
			}

			defer cancel()
			rec, err := elasticsearch.NewRecorder(ctx, log, nil, nil, "d", "nowhere", "d", timeout)
			return ctx, rec, err, nil, nil

		default:
			return nil, nil, nil, nil, nil
		}
	})
}

func TestElasticsearchRecorderConstruction(t *testing.T) {
	recorder.TestRecorderConstruction(t, func(payloadChan chan *recorder.RecordJob, name, indexName string, timeout time.Duration) recorder.DataRecorder {
		log := lib.DiscardLogger()
		endpoint := "http://127.0.0.1:9200"
		rec, _ := elasticsearch.NewRecorder(context.Background(), log, payloadChan, nil, name, endpoint, indexName, timeout)
		destroyIndex(t, endpoint, indexName)
		return rec
	})
}
