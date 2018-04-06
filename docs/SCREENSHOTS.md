# Screenshots

### Table of Contents

1. [Memory Usage](#memory-usage)
    * [Three Apps](#three-apps)
    * [Expipe Memory Usage](#expipe-memory-usage)
    * [Expipe Stack In Use](#expipe-stack-in-use)
    * [Expipe Total Allocations](#expipe-total-allocations)
    * [Expipe Heap Objects](#expipe-heap-objects)
    * [Expipe Heap Allocations](#expipe-heap-allocations)
    * [Expipe Heap In Use](#expipe-heap-in-use)
    * [Expipe Data Collections](#expipe-data-collections)
2. [JSON Tables](#json-tables)

## Memory Usage

#### Three Apps
This graph shows memory stats for two apps plus expipe's own
metrics. The goroutine count comes from expipe as the other two are not exposing
these information.
![Colored](http://i.imgur.com/gTPOCsD.png)

#### Expipe Memory Usage
This is the same graph as above, notice how we used a lucene query to filter
out other apps
![Colored](http://i.imgur.com/6U2hxlp.png)

#### Expipe Stack In Use
![Colored](http://i.imgur.com/F28MWZY.png)

#### Expipe Total Allocations
![Colored](http://i.imgur.com/Tig1k8t.png)

#### Expipe Heap Objects
![Colored](http://i.imgur.com/s8p9br0.png)

#### Expipe Heap Allocations
![Colored](http://i.imgur.com/U6XEqah.png)

#### Expipe Heap In Use
![Colored](http://i.imgur.com/I1yY3kN.png)

#### Expipe Data Collections

![Colored](http://i.imgur.com/XcjlwlB.png)

## JSON Tables

This is an example of how the data is represented inside elasticsearch

![Colored](http://i.imgur.com/waal9cu.png)
