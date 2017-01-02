// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package elasticsearch_test

import (
	"context"
	"errors"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/arsham/expvastic/datatype"
	"github.com/arsham/expvastic/lib"
	"github.com/arsham/expvastic/recorder"
	"github.com/arsham/expvastic/recorder/elasticsearch"
	recorder_testing "github.com/arsham/expvastic/recorder/testing"
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

func setupWithURL(t *testing.T, URL string, indexName string) (ctx context.Context, rec *elasticsearch.Recorder, err error) {
	log := lib.DiscardLogger()
	ctx = context.Background()

	timeout := 10 * time.Millisecond
	if isTravis() {
		timeout = 10 * time.Second
	}

	rec, err = elasticsearch.New(ctx, log, "reader_example", URL, indexName, timeout, 5)
	return ctx, rec, err
}

func setup(t *testing.T) (ctx context.Context, rec *elasticsearch.Recorder, teardown func(), err error) {
	endpoint := "http://127.0.0.1:9200"
	indexName := "expvastic_test"
	ctx, rec, err = setupWithURL(t, endpoint, indexName)
	return ctx, rec, func() {
		destroyIndex(t, endpoint, indexName)
	}, err
}

func TestRecorderCommunication(t *testing.T) {
	recorder_testing.TestRecorderCommunication(t, func(testCase int) (context.Context, recorder.DataRecorder, error, func()) {
		switch testCase {
		case recorder_testing.RecorderReceivesPayloadTestCase:
			ctx, rec, teardown, err := setup(t)
			return ctx, rec, err, teardown

		case recorder_testing.RecorderSendsResultTestCase:
			ctx, rec, teardown, err := setup(t)
			return ctx, rec, err, teardown

		case recorder_testing.RecorderClosesTestCase:
			ctx, rec, teardown, err := setup(t)
			return ctx, rec, err, teardown

		default:
			return nil, nil, nil, nil
		}
	})
}

func TestElasticsearchRecorderConstruction(t *testing.T) {
	recorder_testing.TestRecorderConstruction(t, func(testCase int, name, endpoint, indexName string, timeout time.Duration, backoff int) (recorder.DataRecorder, error) {
		switch testCase {
		case recorder_testing.RecorderConstructionCasesTestCase:
			log := lib.DiscardLogger()
			rec, err := elasticsearch.New(context.Background(), log, name, endpoint, indexName, timeout, backoff)
			destroyIndex(t, endpoint, indexName)
			return rec, err

		case recorder_testing.RecorderErrorsOnInvalidEndpointTestCase:
			log := lib.DiscardLogger()
			return elasticsearch.New(context.Background(), log, name, endpoint, indexName, timeout, backoff)

		default:
			return nil, nil
		}
	})
}

func TestElasticsearchRecorderEndpointManeuvers(t *testing.T) {
	recorder_testing.TestRecorderEndpointManeuvers(t, func(testCase int) (context.Context, recorder.DataRecorder, error, func()) {
		switch testCase {
		case recorder_testing.RecorderErrorsOnUnavailableEndpointTestCase:
			var err error
			log := lib.DiscardLogger()
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

			timeout := 1 * time.Millisecond
			if isTravis() {
				timeout = 10 * time.Second
			}

			rec, err := elasticsearch.New(ctx, log, "d", "http://nowherelocalhost", "d", timeout, 5)
			return ctx, rec, err, func() {
				cancel()
			}

		case recorder_testing.RecorderBacksOffOnEndpointGoneTestCase:
			ctx, rec, teardown, err := setup(t)
			rec.SetRecordFunc(func(ctx context.Context, typeName string, timestamp time.Time, list datatype.DataContainer) error {
				return &url.Error{Op: "GET", URL: "nowhere", Err: errors.New("getsockopt: connection refused")}
			})
			return ctx, rec, err, func() {
				teardown()
			}

		default:
			return nil, nil, nil, nil
		}
	})
}
