package es

import (
	"github.com/olivere/elastic/v7"
	"kumparan/internal/config"
	"log"
	"os"
	"time"
)

var (
	esClient *elastic.Client
)

func InitializeElasticClient(ec config.ElasticConf) (*elastic.Client, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(ec.DSN()),
		elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(10*time.Second),
		elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
		elastic.SetInfoLog(log.New(os.Stdout, "", log.LstdFlags)),
	)
	if err != nil {
		return nil, err
	}

	esClient = client
	return client, err
}

func GetElasticClient() *elastic.Client {
	return esClient
}
