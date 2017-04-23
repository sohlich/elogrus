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

	"reflect"

	"golang.org/x/net/context"
)

type Log struct{}

func (l Log) Printf(format string, args ...interface{}) {
	log.Printf(format+"\n", args)
}

func TestHook(t *testing.T) {
	if r, err := http.Get("http://localhost:7777"); err != nil {
		log.Fatal("Elastic not reachable")
	} else {
		buf, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		fmt.Println(string(buf))
	}

	client, err := elastic.NewClient(elastic.SetTraceLog(Log{}),
		elastic.SetURL("http://localhost:7777"),
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

func TestError(t *testing.T) {
	client, err := elastic.NewClient(elastic.SetTraceLog(Log{}),
		elastic.SetURL("http://localhost:7777"),
		elastic.SetHealthcheck(false),
		elastic.SetSniff(false))

	if err != nil {
		log.Panic(err)
	}

	_, err = client.
		DeleteIndex("errorlog").
		Do(context.TODO())

	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	time.Sleep(1 * time.Second)

	hook, err := NewElasticHook(client, "localhost", logrus.DebugLevel, "errorlog")
	if err != nil {
		log.Panic(err)
		t.FailNow()
	}
	logrus.AddHook(hook)

	logrus.WithError(fmt.Errorf("This is error")).
		Error("Failed to handle invalid api response")

	time.Sleep(1 * time.Second)

	termQuery := elastic.NewTermQuery("Host", "localhost")
	searchResult, err := client.Search().
		Index("errorlog").
		Query(termQuery).
		Do(context.TODO())

	if !(searchResult.Hits.TotalHits >= 1) {
		t.Error("No log created")
		t.FailNow()
	}

	data := searchResult.Each(reflect.TypeOf(logrus.Entry{}))

	for _, d := range data {
		if l, ok := d.(logrus.Entry); ok {
			if errData, ok := l.Data[logrus.ErrorKey]; !ok && errData != "This is error" {
				t.Error("Error not serialized properly")
				t.FailNow()
			}
		}
	}
}
