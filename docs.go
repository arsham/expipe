// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package expvastic can read from any endpoints that provides expvar data and ships them to elasticsearch. You can inspect the metrics with kibana.
//
// Dashboard is provided here: https://github.com/arsham/expvastic/blob/master/bin/dashboard.json
//
// Please refer to golang's expvar documentation for more information.
//
// Here is a couple of screenshots: http://i.imgur.com/6kB88g4.png and  http://i.imgur.com/0ROSWsM.png
//
// Installation guides can be found on github page: https://github.com/arsham/expvastic
//
// At the heart of this package, there is Engine. It acts like a glue between a Reader and a Recorder. Messages are transfered in a package called DataContainer, which is a list of DataType objects.
//
// Here an example configuration, save it somewhere (let's call it expvastic.yml for now):
//
//    settings:
//        debug_evel: info
//
//    readers:
//        FirstApp: # service name
//            type: expvar
//            type_name: my_app1 # this is the elasticsearch type name
//            endpoint: localhost:1234
//            routepath: /debug/vars
//            interval: 500ms
//            timeout: 3s
//            log_level: debug
//            backoff: 10
//        SomeApplication:
//            type: expvar
//            type_name: SomeApplication
//            endpoint: localhost:1235
//            routepath: /debug/vars
//            interval: 500ms
//            timeout: 13s
//            log_level: debug
//            backoff: 10
//
//    recorders:
//        main_elasticsearch:
//            type: elasticsearch
//            endpoint: 127.0.0.1:9200
//            index_name: expvastic
//            timeout: 8s
//            backoff: 10
//        the_other_elasticsearch:
//            type: elasticsearch
//            endpoint: 127.0.0.1:9200
//            index_name: expvastic
//            timeout: 18s
//            backoff: 10
//
//    routes:
//        route1:
//            readers:
//                - FirstApp
//            recorders:
//                - main_elasticsearch
//        route2:
//            readers:
//                - FirstApp
//                - SomeApplication
//            recorders:
//                - main_elasticsearch
//        route3:
//            readers:
//                - SomeApplication
//            recorders:
//                - main_elasticsearch
//                - the_other_elasticsearch
//
// Then run the application:
//
//    expvasyml -c expvastic.yml
//
// You can mix and match the routes, but the engine will choose the best setup to achive your goal without duplicating the results. For instance assume you set the routes like this:
//
//  readers:
//      app_0:
//      app_1:
//      app_2:
//  recorders:
//      elastic_0:
//      elastic_1:
//      elastic_2:
//      elastic_3:
//  routes:
//      route1:
//          readers:
//              - app_0
//              - app_2
//          recorders:
//              - elastic_1
//      route2:
//          readers:
//              - app_0
//          recorders:
//              - elastic_1
//              - elastic_2
//              - elastic_3
//      route2:
//          readers:
//              - app_1
//              - app_2
//          recorders:
//              - elastic_1
//              - elastic_0
//
// Expvastic creates three engines like so:
//
//     Data from app_0 will be shipped to: elastic_1, elastic_2 and elastic_3
//     Data from app_1 will be shipped to: elastic_1 and, elastic_0
//     Data from app_2 will be shipped to: elastic_1 and, elastic_0
//
// For running tests:
//   go test $(glide nv)
//
// For getting test coverages, use this gist: https://gist.github.com/arsham/f45f7e7eea7e18796bc1ed5ced9f9f4a. Then run:
//
//    goverall
//
// For getting benchmarks
//
//   go test $(glide nv) -run=^$ -bench=.
//
// For showing the memory and cpu profiles, on each folder run:
//   BASENAME=$(basename $(pwd))
//   go test -run=^$ -bench=. -cpuprofile=cpu.out -benchmem -memprofile=mem.out
//   go tool pprof -pdf $BASENAME.test cpu.out > cpu.pdf && open cpu.pdf
//   go tool pprof -pdf $BASENAME.test mem.out > mem.pdf && open mem.pdf
//
// Use of this source code is governed by the Apache 2.0 license. License that can be found in the LICENSE file.
package expvastic
