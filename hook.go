package elogrus

import (
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"gopkg.in/olivere/elastic.v3"
)

type ElasticHook struct {
	client *elastic.Client
	host   string
	index  string
	levels []logrus.Level
}

func NewElasticHook(client *elastic.Client, host string, level logrus.Level, index string) *ElasticHook {
	levels := []logrus.Level{}
	for _, l := range []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	} {
		if l <= level {
			levels = append(levels, l)
		}
	}

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(index).Do()
	if err != nil {
		// Handle error
		panic(err)
	}
	if !exists {
		// Create a new index.
		createIndex, err := client.CreateIndex(index).Do()
		if err != nil {
			// Handle error
			panic(err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}
	}

	return &ElasticHook{
		client: client,
		host:   host,
		index:  index,
		levels: levels,
	}
}

func (hook *ElasticHook) Fire(entry *logrus.Entry) error {

	level := entry.Level.String()

	msg := struct {
		Host      string
		Timestamp string
		Message   string
		Data      logrus.Fields
		Level     string
	}{
		hook.host,
		entry.Time.UTC().Format(time.RFC3339Nano),
		entry.Message,
		entry.Data,
		strings.ToUpper(level),
	}

	_, err := hook.client.
		Index().
		Index(hook.index).
		Type("log").
		BodyJson(msg).
		Do()

	return err
}

func (hook *ElasticHook) Levels() []logrus.Level {
	return hook.levels
}
