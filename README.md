# About

This is an early release and it is private. There will be a lot of changes soon but I'm planning to finalise and make it public  as soon as I can. I hope you enjoy it!

Expvastic can read from an endpoint which provides expvar data and ship them to elasticsearch. Please refer to golang's [expvar documentation](https://golang.org/pkg/expvar/) for more information.

Here is a couple of screenshots:

![Colored](http://i.imgur.com/6kB88g4.png)
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
docker run -it --rm --name expvastic --ulimit nofile=98304:98304 -v "/path/to/somewhere/expvastic":/usr/share/elasticsearch/data elasticsearch
docker run -it --rm --name kibana -p 80:5601 --link expvastic:elasticsearch -p 5601:5601 kibana
```

## Running

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

Please note that the name of the app will be changed to expvastic.

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
- [ ] Read from yaml/toml/json configuration files
- [ ] Create different timeouts for each reader/recorder
- [ ] Read from etcd/consul for configurations

## LICENSE

Use of this source code is governed by the Apache 2.0 license. License that can be found in the LICENSE file.

Thanks!
