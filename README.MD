# ElasticSearch Hook for [Logrus](https://github.com/Sirupsen/logrus) <img src="http://i.imgur.com/hTeVwmJ.png" width="40" height="40" alt=":walrus:" class="emoji" title=":walrus:"/>
<img src="https://travis-ci.org/sohlich/elogrus.svg?branch=master" />

## Releases
**Notice that the master branch always refers to the latest version of Elastic. If you want to use stable versions of elogus, you should use the packages released via [gopkg.in](https://gopkg.in).**

Here's the version matrix:

ElasticSearch version | Elastic version  | Package URL                              | Remarks |
----------------------|------------------|------------------------------------------|---------|
7.x                   | 7.0              | [`gopkg.in/sohlich/elogrus.v7`](http://gopkg.in/sohlich/elogrus.v7)| Use Go modules.
6.x                   | 6.0              | [`gopkg.in/sohlich/elogrus.v3`](http://gopkg.in/sohlich/elogrus.v3)| Actively maintained.
5.x                   | 5.0              | [`gopkg.in/sohlich/elogrus.v2`](http://gopkg.in/sohlich/elogrus.v2)| Actively maintained.
2.x                   | 3.0              | [`gopkg.in/sohlich/elogrus.v1`](http://gopkg.in/sohlich/elogrus.v1)| Deprecated. Please update.


## Changelog
- elastic 7.x support (currently in master)


## Import

```
go get gopkg.in/sohlich/elogrus.v3
```

## Usage

```go
package main

import (
	"github.com/sirupsen/logrus.v3"
	"gopkg.in/sohlich/elogrus"
	"gopkg.in/olivere/elastic"
)


func main() {
	log := logrus.New()
	client, err := elastic.NewClient(elastic.SetURL("http://localhost:9200"))
	if err != nil {
		log.Panic(err)
	}	
	hook, err := elogrus.NewElasticHook(client, "localhost", logrus.DebugLevel, "mylog")
	if err != nil {
		log.Panic(err)
	}	
	log.Hooks.Add(hook)

	log.WithFields(logrus.Fields{
		"name": "joe",
		"age":  42,
	}).Error("Hello world!")
}
```

### Asynchronous hook

```go
	...
	elogrus.NewAsyncElasticHook(client, "localhost", logrus.DebugLevel, "mylog")
	...
```
