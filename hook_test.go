package elogrus

import (
	"log"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/olivere/elastic.v3"
)

func TestHook(t *testing.T) {
	client, err := elastic.NewClient(elastic.SetURL("http://localhost:9200"))
	if err != nil {
		log.Panic(err)
	}

	logrus.AddHook(NewElasticHook(client, "localhost", logrus.DebugLevel, "goplag"))

	for index := 0; index < 1000; index++ {
		logrus.Infof("Hustej msg %d", time.Now().Unix())
	}
}
