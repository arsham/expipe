# About

This is an early release and it is under heavy development. There will be a lot of changes soon but I'm planning to finalise the API as soon as I can. I hope you enjoy it!

Expvastic can read from an endpoint which provides expvar data and ship them to elasticsearch. Please refer to golang's [expvar documentation](https://golang.org/pkg/expvar/) for more information.

Here is a couple of screenshots:

![Colored](http://i.imgur.com/83vbwoM.png)
![Colored](http://i.imgur.com/0ROSWsM.png)

## Installing

You need golang 1.7 (I haven't tested it with older versions, but they should be fine) and [glide](https://github.com/Masterminds/glide) installed. Simply do:

```bash
go get github.com/arsham/expvastic/...
cd $GOPATH/src/github.com/arsham/expvastic
glide install
```

You also need elasticsearch and kibana, here is a couple of docker images for you:

```bash
docker run -d --restart always --name expvastic -p 9200:9200 --ulimit nofile=98304:98304 -v "/path/to/somewhere/expvastic":/usr/share/elasticsearch/data elasticsearch
docker run -d --restart always --name kibana -p 80:5601 --link expvastic:elasticsearch -p 5601:5601 kibana
```

## Usage

### With Flags

With this method you can only have one reader and ship to one recorder. Consider the next section for more flexible setup. The defaults are sensible to use, you only need to point the app to two endpoints, and it does the rest for you:

```bash
expvastic -reader="localhost:1234/debug/vars" -recorder="localhost:9200"
```

For more flags run:
```bash
expvastic -h
```

### With Configuration File

Here an example configuration, save it somewhere (let's call it expvastic.yml for now):

```yaml
settings:
    debug_evel: info

readers:
    FirstApp: # service name
        type: expvar
        endpoint: localhost:1234
        routepath: /debug/vars
        interval: 500ms
        timeout: 3s
        log_level: debug
        backoff: 10
    SomeApplication:
        type: expvar
        endpoint: localhost:1235
        routepath: /debug/vars
        interval: 500ms
        timeout: 13s
        log_level: debug
        backoff: 10

recorders:
    main_elasticsearch:
        type: elasticsearch
        endpoint: 127.0.0.1:9200
        index_name: expvastic
        type_name: my_app1
        timeout: 8s
        backoff: 10
    the_other_elasticsearch:
        type: elasticsearch
        endpoint: 127.0.0.1:9200
        index_name: expvastic
        type_name: SomeApplication
        timeout: 18s
        backoff: 10

routes:
    route1:
        readers:
            - FirstApp
        recorders:
            - main_elasticsearch
    route2:
        readers:
            - FirstApp
            - SomeApplication
        recorders:
            - main_elasticsearch
    route3:
        readers:
            - SomeApplication
        recorders:
            - main_elasticsearch
            - the_other_elasticsearch
```

Then run the application:

```bash
expvasyml -c expvastic.yml
```

You can mix and match the routes, but the engine will choose the best setup to achive your goal without duplicating the results. For instance assume you set the routes like this:

```yaml
 readers:
     app_0:
     app_1:
     app_2:
 recorders:
     elastic_0:
     elastic_1:
     elastic_2:
     elastic_3:
 routes:
     route1:
         readers:
             - app_0
             - app_2
         recorders:
             - elastic_1
     route2:
         readers:
             - app_0
         recorders:
             - elastic_1
             - elastic_2
             - elastic_3
     route2:
         readers:
             - app_1
             - app_2
         recorders:
             - elastic_1
             - elastic_0
```

Expvastic creates three engines like so:

```
    Data from app_0 will be shipped to: elastic_1, elastic_2 and elastic_3
    Data from app_1 will be shipped to: elastic_1 and, elastic_0
    Data from app_2 will be shipped to: elastic_1 and, elastic_0
```

## Running Tests

```bash
go test $(glide nv)
```

### Test Coverage

Use this [gist](https://gist.github.com/arsham/f45f7e7eea7e18796bc1ed5ced9f9f4a). Then run the function, that's all!

## TODO
- [ ] Decide how to show GC information correctly
- [ ] When reader/recorder are not available, don't check right away
- [ ] Create UUID for messages in order to log them
- [ ] Read from multiple sources
- [ ] Record expvastic's own metrics
- [ ] Use dates on index names
- [ ] Read from other providers; python, JMX etc.
- [ ] Read from log files
- [ ] Benchmarks
- [ ] Create a docker image
- [ ] Make a compose file
- [ ] Gracefully shutdown
- [ ] Share kibana setups
- [=] Read from yaml/toml/json configuration files
- [X] Create different timeouts for each reader/recorder
- [ ] Read from etcd/consul for configurations

## LICENSE

Use of this source code is governed by the Apache 2.0 license. License that can be found in the LICENSE file.

Thanks!
