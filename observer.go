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

// observer contains two maps for traking the recorders.
type observer struct {
    logger      logrus.FieldLogger
    payloadChan chan<- *reader.ReadJobResult
    removeChan  chan string
    // TODO: add strike, if a recorder gets 5 strikes, remove them from the list

    rmu       sync.RWMutex
    recorders map[string]recorder.DataRecorder
    dmu       sync.RWMutex
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
            o.rmu.RLock()
            if len(o.recorders) == 0 {
                // we need to return when the last item is removed.
                o.rmu.RUnlock()
                return
            }
            o.rmu.RUnlock()
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
    doneChan := recorder.Start(ctx)

    o.rmu.Lock()
    o.recorders[recorder.Name()] = recorder
    o.rmu.Unlock()

    o.dmu.Lock()
    o.doneChans[recorder.Name()] = doneChan
    o.dmu.Unlock()

    go func() {

        o.dmu.RLock()
        doneCh := o.doneChans[recorder.Name()]
        o.dmu.RUnlock()

        select {
        case <-doneCh:
            // the recorder has closed its done channel
            o.removeChan <- recorder.Name()
        case <-ctx.Done():
            // the parent context is canceled
            return
        }
    }()
}

func (o *observer) Remove(name string) {
    o.rmu.Lock()
    defer o.rmu.Unlock()
    o.dmu.Lock()
    defer o.dmu.Unlock()

    delete(o.recorders, name)
    delete(o.doneChans, name)
}

// Send distributes the payload to the recorders.
// It creates a separate goroutine for each recorde.
// It will remove them from the map when they close their done channel.
func (o *observer) Send(ctx context.Context, typeName string, t time.Time, payload datatype.DataContainer) {
    o.rmu.RLock()
    recorders := o.recorders
    o.rmu.RUnlock()

    for name, rec := range recorders {
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
                // job was sent, let's reset the timer for getting the error message.
                if !timer.Stop() {
                    <-timer.C
                }
                timer.Reset(timeout)
            case <-timer.C:
                o.logger.Warn("timedout before receiving the error")
                drainErrorChan(timer, timeout, errChan)
            case <-ctx.Done():
                o.logger.Warnf("main context was closed before receiving the error response: %s", ctx.Err().Error())
                drainErrorChan(timer, timeout, errChan)
            case <-o.doneChans[name]:
                o.logger.Warnf("recorder %s just opted out", name)
                o.removeChan <- name
                // the recorder has quit, there is no point reading the error result
                // let's just drain it to be sure.
                drainErrorChan(timer, timeout, errChan)
                return
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
                drainErrorChan(timer, timeout, errChan)
            case <-o.doneChans[name]:
                o.logger.Warnf("recorder %s just opted out", name)
                o.removeChan <- name
                drainErrorChan(timer, timeout, errChan)
            }
        }(name, rec)
    }
}

// Drains the error channel to make sure we are not leaving anything in memory
// QUESTION: what happens if we don't drain this channel?
func drainErrorChan(timer *time.Timer, timeout time.Duration, errChan chan error) {
    go func() {
        if !timer.Stop() {
            <-timer.C
        }
        timer.Reset(timeout)
        select {
        case <-errChan:
        case <-timer.C:
        }
    }()
}
