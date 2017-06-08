package elogrus

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"gopkg.in/olivere/elastic.v5"
)

var (
	// Fired if the
	// index is not created
	ErrCannotCreateIndex = fmt.Errorf("Cannot create index")
)

type IndexNameFunc func() string

// ElasticHook is a logrus
// hook for ElasticSearch
type ElasticHook struct {
	client    *elastic.Client
	host      string
	index     IndexNameFunc
	levels    []logrus.Level
	ctx       context.Context
	ctxCancel context.CancelFunc
}

// NewElasticHook creates new hook
// client - ElasticSearch client using gopkg.in/olivere/elastic.v5
// host - host of system
// level - log level
// index - name of the index in ElasticSearch
func NewElasticHook(client *elastic.Client, host string, level logrus.Level, index string) (*ElasticHook, error) {
	return NewElasticHookWithFunc(client, host, level, func() string { return index })
}

// NewElasticHookWithFunc creates new hook with
// function that provides the index name. This is useful if the index name is
// somehow dynamic especially based on time.
// client - ElasticSearch client using gopkg.in/olivere/elastic.v5
// host - host of system
// level - log level
// indexFunc - function providing the name of index
func NewElasticHookWithFunc(client *elastic.Client, host string, level logrus.Level, indexFunc IndexNameFunc) (*ElasticHook, error) {
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

	ctx, cancel := context.WithCancel(context.TODO())

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(indexFunc()).Do(ctx)
	if err != nil {
		// Handle error
		return nil, err
	}
	if !exists {
		createIndex, err := client.CreateIndex(indexFunc()).Do(ctx)
		if err != nil {
			return nil, err
		}
		if !createIndex.Acknowledged {
			return nil, ErrCannotCreateIndex
		}
	}

	return &ElasticHook{
		client:    client,
		host:      host,
		index:     indexFunc,
		levels:    levels,
		ctx:       ctx,
		ctxCancel: cancel,
	}, nil
}

// Fire is required to implement
// Logrus hook
func (hook *ElasticHook) Fire(entry *logrus.Entry) error {

	level := entry.Level.String()

	if e, ok := entry.Data[logrus.ErrorKey]; ok && e != nil {
		if err, ok := e.(error); ok {
			entry.Data[logrus.ErrorKey] = err.Error()
		}
	}

	msg := struct {
		Host      string
		Timestamp string `json:"@timestamp"`
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
		Index(hook.index()).
		Type("log").
		BodyJson(msg).
		Do(hook.ctx)

	return err
}

// Required for logrus
// hook implementation
func (hook *ElasticHook) Levels() []logrus.Level {
	return hook.levels
}

// Cancels all calls to
// elastic
func (hook *ElasticHook) Cancel() {
	hook.ctxCancel()
}
