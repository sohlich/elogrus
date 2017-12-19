package elogrus

import (
	"context"
	"log"
	"net/http"
	"testing"
	"time"

	"fmt"

	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gopkg.in/olivere/elastic.v5"

	"reflect"
)

//docker run -it --rm -p 7777:9200 -p 5601:5601 sebp/elk

type NewHookFunc func(client *elastic.Client, host string, level logrus.Level, index string) (*ElasticHook, error)

type Log struct{}

func (l Log) Printf(format string, args ...interface{}) {
	log.Printf(format+"\n", args)
}

func TestSyncHook(t *testing.T) {
	hookTest(NewElasticHook, t)
}

func TestAsyncHook(t *testing.T) {
	hookTest(NewAsyncElasticHook, t)
}

func hookTest(hookfunc NewHookFunc, t *testing.T) {
	if r, err := http.Get("http://127.0.0.1:7777"); err != nil {
		log.Fatal("Elastic not reachable")
	} else {
		buf, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		fmt.Println(string(buf))
	}

	client, err := elastic.NewClient(elastic.SetTraceLog(Log{}),
		elastic.SetURL("http://127.0.0.1:7777"),
		elastic.SetHealthcheck(false),
		elastic.SetSniff(false))

	if err != nil {
		log.Panic(err)
	}

	client.
		DeleteIndex("goplag").
		Do(context.TODO())

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
		t.Errorf("Not all logs pushed to elastic: expected %d got %d", 100, searchResult.Hits.TotalHits)
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

	client.
		DeleteIndex("errorlog").
		Do(context.TODO())

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
