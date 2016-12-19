// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

// func ExampleEngine_sendJob() {
//     ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
//     defer ts.Close()

//     readerChan := make(chan struct{})
//     conf := simpleRecorderSetup(ts.URL, readerChan, nil) // url, reader channel, recorder channel

//     ctx, cancel := context.WithCancel(context.Background())
//     cl := expvastic.NewEngine(ctx, conf)
//     go cl.Start()

//     select {
//     case <-ctx.Done():
//         panic("job wasn't sent")
//     case j := <-readerChan:
//         fmt.Println("job was sent successfully")
//         fmt.Printf("Job value: %v\n", j)
//         fmt.Printf("j == struct{}{}: %t\n", j == struct{}{})
//         cancel()
//     }

//     <-ctx.Done()

//     // Output:
//     // job was sent successfully
//     // Job value: {}
//     // j == struct{}{}: true
// }

// func ExampleEngine_RecorderReturnsResult() {
//     ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
//     defer ts.Close()

//     readerChan := make(chan struct{})
//     recJobChan := make(chan *recorder.RecordJob)
//     resultChan := make(chan *reader.ReadJobResult)

//     conf := simpleRecReaderSetup(ts.URL, readerChan, recJobChan, resultChan) // url, reader channel, job recorder channel, result channel

//     ctx, cancel := context.WithCancel(context.Background())
//     cl := expvastic.NewEngine(ctx, *conf)
//     go cl.Start()
//     ftype := datatype.FloatType{"test", 666.66}
//     ftypeStr := fmt.Sprintf("{%s}", ftype)

//     msg := ioutil.NopCloser(strings.NewReader(ftypeStr))
//     resultChan <- &reader.ReadJobResult{Res: msg}
//     r := <-recJobChan
//     result := r.Payload.List()[0]
//     fmt.Println(result.String())
//     fmt.Printf("Result is sent value: %t\n", result.String() == ftype.String())
//     cancel()
//     <-ctx.Done()
//     // Output:
//     // "test":666.660000
//     // Result is sent value: true
// }
