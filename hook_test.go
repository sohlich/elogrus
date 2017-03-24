package elogrus

import (
	"log"
	"net/http"
	"testing"
	"time"

	"fmt"

	"io/ioutil"

	"github.com/Sirupsen/logrus"
	"gopkg.in/olivere/elastic.v5"

	"golang.org/x/net/context"
)

type Log struct{}

func (l Log) Printf(format string, args ...interface{}) {
	log.Printf(format+"\n", args)
}

func TestHook(t *testing.T) {
	if r, err := http.Get("http://localhost:9200"); err != nil {
		log.Fatal("Elastic not reachable")
	} else {
		buf, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		fmt.Println(string(buf))
	}

	client, err := elastic.NewClient(elastic.SetTraceLog(Log{}),
		elastic.SetURL("http://localhost:9200"),
		elastic.SetHealthcheck(false),
		elastic.SetSniff(false))

	if err != nil {
		log.Panic(err)
	}

	hook, err := NewElasticHook(client, "localhost", logrus.DebugLevel, "goplag")
	if err != nil {
		log.Panic(err)
		t.FailNow()
	}
	logrus.AddHook(hook)

	for index := 0; index < 100; index++ {
		logrus.Infof("Hustej msg %d", time.Now().Unix())
	}

	time.Sleep(5 * time.Second)

	termQuery := elastic.NewTermQuery("Host", "localhost")
	searchResult, err := client.Search().
		Index("goplag").
		Query(termQuery).
		Do(context.TODO())

	if searchResult.Hits.TotalHits != 100 {
		t.Error("Not all logs pushed to elastic")
		t.FailNow()
	}
}
