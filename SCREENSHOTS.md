# Screenshots

### Table of Contents

1. [Memory Usage](#memory-usage)
  * [Three Apps](#three-apps)
  * [Expvastic Memory Usage](#expvastic-memory-usage)
  * [Expvastic Stack In Use](#expvastic-stack-in-use)
  * [Expvastic Total Allocations](#expvastic-total-allocations)
  * [Expvastic Heap Objects](#expvastic-heap-objects)
  * [Expvastic Heap Allocations](#expvastic-heap-allocations)
  * [Expvastic Heap In Use](#expvastic-heap-in-use)
  * [Expvastic Data Collections](#expvastic-data-collections)
2. [JSON Tables](#json-tables)

## Memory Usage

#### Three Apps
This graph shows memory stats for two apps plus expvastic's own metrics. The goroutine count comes from expvastic as the other two are not exposing these information.
![Colored](http://i.imgur.com/gTPOCsD.png)

#### Expvastic Memory Usage
This is the same graph as above, notice how we used a lucene query to filter out other apps
![Colored](http://i.imgur.com/6U2hxlp.png)

#### Expvastic Stack In Use
![Colored](http://i.imgur.com/F28MWZY.png)

#### Expvastic Total Allocations
![Colored](http://i.imgur.com/Tig1k8t.png)

#### Expvastic Heap Objects
![Colored](http://i.imgur.com/s8p9br0.png)

#### Expvastic Heap Allocations
![Colored](http://i.imgur.com/U6XEqah.png)

#### Expvastic Heap In Use
![Colored](http://i.imgur.com/I1yY3kN.png)

#### Expvastic Data Collections

![Colored](http://i.imgur.com/XcjlwlB.png)

## JSON Tables

This is an example of how the data is represented inside elasticsearch

![Colored](http://i.imgur.com/waal9cu.png)


