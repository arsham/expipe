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
// Use of this source code is governed by the Apache 2.0 license. License that can be found in the LICENSE file.
package expvastic
