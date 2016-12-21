// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
    "context"
    "time"

    "sync"

    "github.com/Sirupsen/logrus"
    "github.com/arsham/expvastic/datatype"
    "github.com/arsham/expvastic/reader"
    "github.com/arsham/expvastic/recorder"
)

type observer struct {
    logger      logrus.FieldLogger
    payloadChan chan<- *reader.ReadJobResult
    removeChan  chan string
    // TODO: add strike, if a recorder gets 5 strikes, remove them from the list

    mu        sync.RWMutex
    recorders map[string]recorder.DataRecorder
    doneChans map[string]chan struct{}
}

func newobserver(ctx context.Context, logger logrus.FieldLogger, initialLen int) *observer {
    o := &observer{
        logger:      logger,
        payloadChan: make(chan<- *reader.ReadJobResult, initialLen), // TODO
        recorders:   make(map[string]recorder.DataRecorder, initialLen),
        doneChans:   make(map[string]chan struct{}, initialLen),
        removeChan:  make(chan string, initialLen),
    }
    go func() {
        for {
            if len(o.recorders) == 0 {
                return
            }
            select {
            case name := <-o.removeChan:
                o.Remove(name)
            case <-ctx.Done():
                return
            }
        }
    }()
    return o
}

func (o *observer) Add(ctx context.Context, recorder recorder.DataRecorder) {
    o.mu.Lock()
    defer o.mu.Unlock()
    o.recorders[recorder.Name()] = recorder
    o.doneChans[recorder.Name()] = recorder.Start(ctx)
    go func() {
        select {
        case <-o.doneChans[recorder.Name()]:
            o.removeChan <- recorder.Name()
        case <-ctx.Done():
            return
        }
    }()
}

func (o *observer) Remove(name string) {
    o.mu.Lock()
    defer o.mu.Unlock()
    delete(o.recorders, name)
    delete(o.doneChans, name)
}

func (o *observer) Send(ctx context.Context, typeName string, t time.Time, payload datatype.DataContainer) {
    // o.mu.RLock()
    // defer o.mu.RUnlock()

    for name, rec := range o.recorders {
        // we are sending the payload for each recorder separately.
        go func(name string, rec recorder.DataRecorder) {
            o.logger.Debugf("sending payload to %s", name)
            errChan := make(chan error)
            timeout := rec.Timeout() + time.Duration(10*time.Second)
            timer := time.NewTimer(timeout)
            payload := &recorder.RecordJob{
                Ctx:       ctx,
                Payload:   payload,
                IndexName: rec.IndexName(),
                TypeName:  typeName,
                Time:      t,
                Err:       errChan,
            }

            // sending payload
            select {
            case rec.PayloadChan() <- payload:
                // job was sent, let's do the same for the error message.
                if !timer.Stop() {
                    <-timer.C
                }
                timer.Reset(timeout)
            case <-timer.C:
                o.logger.Warn("timedout before receiving the error")
            case <-ctx.Done():
                o.logger.Warnf("main context was closed before receiving the error response: %s", ctx.Err().Error())
            case <-o.doneChans[name]:
                o.logger.Warnf("recorder %s just opted out", name)
                o.removeChan <- name
            }

            // waiting for the result
            select {
            case err := <-errChan:
                if err != nil {
                    o.logger.Errorf("%s", err.Error())
                }
                // received the response
                timer.Stop()
            case <-timer.C:
                o.logger.Warn("timedout before receiving the error")
            case <-ctx.Done():
                o.logger.Warnf("main context was canceled before receiving the error: %s", ctx.Err().Error())
            case <-o.doneChans[name]:
                o.logger.Warnf("recorder %s just opted out", name)
                o.removeChan <- name
            }
        }(name, rec)
    }
}
