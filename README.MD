# ElasticSearch Hook for [Logrus](https://github.com/Sirupsen/logrus) <img src="http://i.imgur.com/hTeVwmJ.png" width="40" height="40" alt=":walrus:" class="emoji" title=":walrus:"/>


## Usage

```go
package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/sohlich/elogrus"
	"gopkg.in/olivere/elastic.v3"
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
