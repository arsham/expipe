# Expvastic

[![Build Status](https://travis-ci.org/arsham/expvastic.svg?branch=master)](https://travis-ci.org/arsham/expvastic)

Expvastic can record your application's `metrics` in [ElasticSearch][elasticsearch] and you can view them with [kibana][kibana]. It can read from any applications (written in any language) that provides metrics in `json` format.

1. [Features](#features)
2. [Installation](#installation)
3. [Kibana](#kibana)
  * [Importing Dashboard](#importing-dashboard)
4. [Usage](#usage)
  * [With Flags](#with-flags)
  * [Advanced](#advanced)
5. [LICENSE](#license)

## Features

* Very lightweight and fast.
* Can read from multiple input.
* Can ship the metrics to multiple databases.
* Shows memory usages and GC pauses of the apps.
* Metrics can be aggregated for different apps (with elasticsearch's type system).
* A kibana dashboard is also provided [here](./bin/dashboard.json).
* Maps values how you define them. For example you can change bytes to megabytes.
* Benchmarks are included.

There are TODO items in the issue section. Feature requests welcome!


Please refer to golang's [expvar documentation][expvar] for more information.

Screenshots can be found in [this](./SCREENSHOTS.md) document. Here is an example:

![Colored](http://i.imgur.com/34kdQe8.png)

## Installation

I will provide a docker image soon, but for now it needs to be installed. You need golang 1.7 and [glide][glide] installed. Simply do:

```bash
go get github.com/arsham/expvastic
cd $GOPATH/src/github.com/arsham/expvastic
glide install
go install ./cmd/expvastic
```

You also need elasticsearch and kibana, here is a couple of docker images you can start with:

```bash
docker run -d --restart always --name expvastic -p 9200:9200 --ulimit nofile=98304:98304 -v "/path/to/somewhere/expvastic":/usr/share/elasticsearch/data elasticsearch
docker run -d --restart always --name kibana -p 80:5601 --link expvastic:elasticsearch -p 5601:5601 kibana
```

## Kibana

Access [the dashboard](http://localhost) (or any other ports you have exposed kibana to, notice the `-p:80:5601` above), and enter `expvastic` as `Index name or pattern` in `management` section.

Select `@timestamp` for `Time-field name`. In case it doesn't show up, click `Index contains time-based events` twice, it will provide you with the timestamp. Then click on create button.

### Importing Dashboard

Go to `Saved Objects` section of `management`, and click on the `import` button. Upload [this](./bin/dashboard.json) file and you're done!

One of the provided dashboards shows the expvastic's own metrics, and you can use the other one for everything you have defined in the configuration file.

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

### Advanced

Please refer to [this](./RECIPES.md) document for advanced configuration and mappings.

## LICENSE

Use of this source code is governed by the Apache 2.0 license. License that can be found in the [LICENSE](./LICENSE) file.

`Enjoy!`


[expvar]: https://golang.org/pkg/expvar/
[glide]: https://github.com/Masterminds/glide
[elasticsearch]: https://github.com/elastic/elasticsearch
[kibana]: https://github.com/elastic/kibana
