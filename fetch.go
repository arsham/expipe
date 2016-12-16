// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/Sirupsen/logrus"
)

type jobResult struct {
    time time.Time
    res  io.ReadCloser
    err  error
}

// Because the caller is reading the resp.Body, it is its job to close it
func fetch(log logrus.FieldLogger, target string, jobs <-chan context.Context, resCh chan jobResult) (done chan struct{}) {
    done = make(chan struct{})
    go func() {
        for job := range jobs {
            r := jobResult{}
            resp, err := getRequest(job, target)
            if err != nil {
                r.err = fmt.Errorf("making request: %s", err)
                resCh <- r
                continue
            }
            r.time = time.Now() // It is sensible to record the time now
            r.res = resp.Body
            resCh <- r
        }
        close(done)
    }()
    return done
}

func getRequest(ctx context.Context, url string) (*http.Response, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    req = req.WithContext(ctx)
    return http.DefaultClient.Do(req)
}
