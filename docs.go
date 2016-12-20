// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package expvastic can read from an endpoint which provides expvar data and ship them to elasticsearch. Please refer to golang's [expvar documentation](https://golang.org/pkg/expvar/) for more information.
// This is an early release and it is private. There will be a lot of changes soon but I'm planning to finalise and make it public  as soon as I can. I hope you enjoy it!
//
//
// Here is a couple of screenshots: http://i.imgur.com/6kB88g4.png and  http://i.imgur.com/0ROSWsM.png
//
// You need golang 1.7 (I haven't tested it with older versions, but they should be fine) and [glide](https://github.com/Masterminds/glide) installed. Simply do:
//
//    go get github.com/arsham/expvastic/...
//    cd $GOPATH/src/github.com/arsham/expvastic
//    glide install
//
// You also need elasticsearch and kibana, here is a couple of docker images for you:
//
//    docker run -it --rm --name expvastic --ulimit nofile=98304:98304 -v "/path/to/somewhere/expvastic":/usr/share/elasticsearch/data elasticsearch
//    docker run -it --rm --name kibana -p 80:5601 --link expvastic:elasticsearch -p 5601:5601 kibana
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
//            endpoint: localhost:1234
//            routepath: /debug/vars
//            interval: 500ms
//            timeout: 3s
//            log_level: debug
//            backoff: 10
//        SomeApplication:
//            type: expvar
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
//            type_name: my_app1
//            timeout: 8s
//            backoff: 10
//        the_other_elasticsearch:
//            type: elasticsearch
//            endpoint: 127.0.0.1:9200
//            index_name: expvastic
//            type_name: SomeApplication
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
// Please note that the name of the app will be changed to expvastic.
//
// For running tests, do the following:
//   go test $(glide nv)
// Use of this source code is governed by the Apache 2.0 license. License that can be found in the LICENSE file.
package expvastic
