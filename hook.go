package elogrus

import (
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"gopkg.in/olivere/elastic.v3"
)

var (
	// Fired if the
	// index is not created
	ErrCannotCreateIndex = fmt.Errorf("Cannot create index")
)

// ElasticHook is a logrus
// hook for ElasticSearch
type ElasticHook struct {
	client *elastic.Client
	host   string
	index  string
	levels []logrus.Level
}

// NewElasticHook creates new hook
// client - ElasticSearch client using gopkg.in/olivere/elastic.v3
// host - host of system
// level - log level
// index - name of the index in ElasticSearch
func NewElasticHook(client *elastic.Client, host string, level logrus.Level, index string) (*ElasticHook, error) {
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
		return nil, err
	}
	if !exists {
		createIndex, err := client.CreateIndex(index).Do()
		if err != nil {
			return nil, err
		}
		if !createIndex.Acknowledged {
			return nil, ErrCannotCreateIndex
		}
	}

	return &ElasticHook{
		client: client,
		host:   host,
		index:  index,
		levels: levels,
	}, nil
}

// Fire is required to implement
// Logrus hook
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

// Required for logrus
// hook implementation
func (hook *ElasticHook) Levels() []logrus.Level {
	return hook.levels
}
