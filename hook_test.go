package elogrus

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

type NewHookFunc func(client *elastic.Client, host string, level logrus.Level, index string) (*ElasticHook, error)

func TestSyncHook(t *testing.T) {
	hookTest(NewElasticHook, "sync-log", t)
}

func TestAsyncHook(t *testing.T) {
	hookTest(NewAsyncElasticHook, "async-log", t)
}

func TestBulkProcessorHook(t *testing.T) {
	hookTest(NewBulkProcessorElasticHook, "bulk-log", t)
}

func hookTest(hookfunc NewHookFunc, indexName string, t *testing.T) {
	if r, err := http.Get("http://127.0.0.1:7777"); err != nil {
		log.Fatal("Elastic not reachable")
	} else {
		buf, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		fmt.Println(string(buf))
	}

	client, err := elastic.NewClient(
		elastic.SetURL("http://127.0.0.1:7777"),
		elastic.SetHealthcheck(false),
		elastic.SetSniff(false))

	if err != nil {
		log.Panic(err)
	}

	client.
		DeleteIndex(indexName).
		Do(context.TODO())

	hook, err := hookfunc(client, "localhost", logrus.DebugLevel, indexName)
	if err != nil {
		log.Panic(err)
	}
	logrus.AddHook(hook)

	samples := 100
	for index := 0; index < samples; index++ {
		logrus.Infof("Hustej msg %d", time.Now().Unix())
	}

	// Allow time for data to be processed.
	time.Sleep(2 * time.Second)

	termQuery := elastic.NewTermQuery("Host", "localhost")
	searchResult, err := client.Search().
		Index(indexName).
		Query(termQuery).
		Do(context.TODO())

	if err != nil {
		t.Errorf("Search error: %v", err)
		t.FailNow()
	}

	if searchResult.TotalHits() != int64(samples) {
		t.Errorf("Not all logs pushed to elastic: expected %d got %d", samples, searchResult.TotalHits())
		t.FailNow()
	}
}

func TestError(t *testing.T) {
	client, err := elastic.NewClient(
		elastic.SetURL("http://localhost:7777"),
		elastic.SetHealthcheck(false),
		elastic.SetSniff(false))

	if err != nil {
		log.Panic(err)
	}

	client.
		DeleteIndex("errorlog").
		Do(context.TODO())

	hook, err := NewElasticHook(client, "localhost", logrus.DebugLevel, "errorlog")
	if err != nil {
		log.Panic(err)
		t.FailNow()
	}
	logrus.AddHook(hook)

	logrus.WithError(fmt.Errorf("this is error")).
		Error("Failed to handle invalid api response")

	// Allow time for data to be processed.
	time.Sleep(1 * time.Second)
	termQuery := elastic.NewTermQuery("Host", "localhost")
	searchResult, err := client.Search().
		Index("errorlog").
		Query(termQuery).
		Do(context.TODO())

	if !(searchResult.TotalHits() >= int64(1)) {
		t.Error("No log created")
		t.FailNow()
	}

	data := searchResult.Each(reflect.TypeOf(logrus.Entry{}))
	for _, d := range data {
		if l, ok := d.(logrus.Entry); ok {
			if errData, ok := l.Data[logrus.ErrorKey]; !ok && errData != "this is error" {
				t.Error("Error not serialized properly")
				t.FailNow()
			}
		}
	}
}
