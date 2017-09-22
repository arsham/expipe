# Reccipes
## Table of Contents


1. [Kibana](#kibana)
    * [Per Application Setup](#per-application-setup)
3. [Configuration File](#configuration-file)
    * [How Routes Are Defined](#how-routes-are-defined)
    * [Mappings](#mappings)
4. [Testing](#testing)
5. [Coverage](#coverage)
6. [Benchmarks](#benchmarks)

## Kibana

### Per Application Setup

On any dashboards/visualisations, you can specify the metrics for a specific app. Let's assume one of your app name (type_name in the configuration file) is called `Arsham`. Then in the search bar on top type in: `_type:Arsham`

## Configuration File

Here an example configuration, save it somewhere (let's call it expvastic.yml for now):

```yaml
settings:
    log_level: info

readers:                           # You can specify the applications you want to show the metrics
    FirstApp:                      # service name
        type: expvar               # the type of reader. More to come soon!
        type_name: AppVastic       # this will be the _type in elasticsearch
        endpoint: localhost:1234   # where the application
        routepath: /debug/vars     # the endpoint that app provides the metrics
        interval: 500ms            # every half a second, it will collect the metrics.
        timeout: 3s                # in 3 seconds it gives in if the application is not responsive
        backoff: 10                # after 10 times the application didn't response, it will stop reading from it
    AnotherApplication:
        type: expvar
        type_name: this_is_awesome
        endpoint: localhost:1235
        routepath: /metrics
        interval: 500ms
        timeout: 13s
        backoff: 10

recorders:                         # This section is where the data will be shipped to
    main_elasticsearch:
        type: elasticsearch        # the type of recorder. More to come soon!
        endpoint: 127.0.0.1:9200
        index_name: expvastic
        timeout: 8s
        backoff: 10
    the_other_elasticsearch:
        type: elasticsearch
        endpoint: 127.0.0.1:9201
        index_name: expvastic
        timeout: 18s
        backoff: 10

routes:                            # You can specify metrics of which application will be recorded in which target
    route1:
        readers:
            - FirstApp
        recorders:
            - main_elasticsearch
    route2:
        readers:
            - FirstApp
            - AnotherApplication
        recorders:
            - main_elasticsearch
    route3:                      # Yes, you can have multiple!
        readers:
            - AnotherApplication
        recorders:
            - main_elasticsearch
            - the_other_elasticsearch
```

Then run the application and point it to the file:

```bash
expvastic -c expvastic.yml
```

### How Routes Are Defined

You can mix and match the routes, but the engine will choose the best setup to achieve your goal without duplicating the results. For instance assume you set the routes like this:

```yaml
readers:
    app_0: type: expvar
    app_1: type: expvar
    app_2: type: expvar
    app_3: type: expvar
    app_4: type: expvar
    app_5: type: expvar
    not_used_app: type: expvar # note that this one is not specified in the routes, therefore it is ignored
recorders:
    elastic_0: type: elasticsearch
    elastic_1: type: elasticsearch
    elastic_2: type: elasticsearch
    elastic_3: type: elasticsearch
routes:
    route1:
        readers:
            - app_0
            - app_2
            - app_4
        recorders:
            - elastic_1
    route2:
        readers:
            - app_0
            - app_5
        recorders:
            - elastic_2
            - elastic_3
    route3:
        readers:
            - app_1
            - app_2
        recorders:
            - elastic_0
            - elastic_1
```
Expvastic creates three engines like so:

```
    elastic_0 records data from app_0, app_1
    elastic_1 records data from app_0, app_1, app_2, app_4
    elastic_2 records data from app_0, app_5
    elastic_3 records data from app_0, app_5
```

### Mappings

You can change the numbers to your liking:

```
gc_types:                      # These inputs will be collected into one list and zero values will be removed
    memstats.PauseEnd
    memstats.PauseNs

memory_bytes:                   # These values will be transoformed from bytes
    StackInuse: mb              # To MB
    memstats.Alloc: gb          # To GB

```

## Testing

To run the tests for the codes, in the root of the application run:

```bash
go test $(glide nv)
```

Or for testing readers:

```bash
go test ./readers
```

## Coverage

Use this [gist](https://gist.github.com/arsham/f45f7e7eea7e18796bc1ed5ced9f9f4a). Then run:

```bash
goverall
```

It will open a browser tab and show you the coverage.

## Benchmarks

To run all benchmarks:

```bash
go test $(glide nv) -run=^$ -bench=.
```

For showing the memory and cpu profiles, on each folder run:

```bash
BASENAME=$(basename $(pwd))
go test -run=^$ -bench=. -cpuprofile=cpu.out -benchmem -memprofile=mem.out
go tool pprof -pdf $BASENAME.test cpu.out > cpu.pdf && open cpu.pdf
go tool pprof -pdf $BASENAME.test mem.out > mem.pdf && open mem.pdf
```
