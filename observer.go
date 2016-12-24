// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic

import (
    "context"
    "expvar"
    "time"

    "sync"

    "github.com/Sirupsen/logrus"
    "github.com/arsham/expvastic/communication"
    "github.com/arsham/expvastic/datatype"
    "github.com/arsham/expvastic/reader"
    "github.com/arsham/expvastic/recorder"
)

var recordsDistributed = expvar.NewInt("Records Distributed")

// observer distributes record jobs to all recorders. It contains two maps for traking the recorders.
type observer struct {
    logger      logrus.FieldLogger
    payloadChan chan<- *reader.ReadJobResult
    resultChan  chan<- error
    removeChan  chan string // when receives a name, will remove the recorder
    // TODO: add strike, if a recorder gets 5 strikes, remove them from the list

    rmu       sync.RWMutex
    recorders map[string]recorder.DataRecorder // map of recorders name->objects

    dmu       sync.RWMutex
    doneChans map[string]<-chan struct{} // map of done channels name->done channel
}

func newobserver(ctx context.Context, logger logrus.FieldLogger, resultChan chan<- error, initialLen int) *observer {
    o := &observer{
        logger:      logger,
        payloadChan: make(chan<- *reader.ReadJobResult, initialLen), // TODO
        recorders:   make(map[string]recorder.DataRecorder, initialLen),
        doneChans:   make(map[string]<-chan struct{}, initialLen),
        removeChan:  make(chan string, initialLen),
        resultChan:  resultChan,
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
                o.unsubscribe(name)
            case <-ctx.Done():
                return
            }
        }
    }()
    return o
}

func (o *observer) subscribe(ctx context.Context, recorder recorder.DataRecorder) {
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

func (o *observer) unsubscribe(name string) {
    o.rmu.Lock()
    defer o.rmu.Unlock()
    o.dmu.Lock()
    defer o.dmu.Unlock()

    delete(o.recorders, name)
    delete(o.doneChans, name)
}

// publish distributes the payload to the recorders.
// It creates a separate goroutine for each recorde.
// It will remove them from the map when they close their done channel.
func (o *observer) publish(ctx context.Context, id communication.JobID, typeName string, t time.Time, payload datatype.DataContainer) {
    o.rmu.RLock()
    recorders := o.recorders
    o.rmu.RUnlock()
    for name, rec := range recorders {
        // we are sending the payload for each recorder separately.
        go func(name string, rec recorder.DataRecorder) {
            o.logger.Debugf("sending payload to %s", name)
            o.dmu.RLock()
            doneChan := o.doneChans[name]
            o.dmu.RUnlock()

            timeout := rec.Timeout() + time.Duration(10*time.Second)
            timer := time.NewTimer(timeout)
            payload := &recorder.RecordJob{
                ID:        id,
                Ctx:       ctx,
                Payload:   payload,
                IndexName: rec.IndexName(),
                TypeName:  typeName,
                Time:      t,
            }
            recordsDistributed.Add(1)

            // sending payload
            select {
            case rec.PayloadChan() <- payload:
                // job was sent, let's reset the timer for getting the error message.
                if !timer.Stop() {
                    <-timer.C
                }

            case <-timer.C:
                o.logger.Warn("timedout before receiving the error")

            case <-ctx.Done():
                o.logger.Warnf("main context was closed before receiving the error response: %s", ctx.Err().Error())
                if !timer.Stop() {
                    <-timer.C
                }

            case <-doneChan:
                o.logger.Warnf("recorder %s just opted out", name)
                o.removeChan <- name
                if !timer.Stop() {
                    <-timer.C
                }
                return
            }
        }(name, rec)
    }
}
