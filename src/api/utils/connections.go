package utils

import (
	"api/models"

	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	es7 "github.com/elastic/go-elasticsearch"
)

func CreateEsConnection(cfg *models.Config) *es7.Client {
	var (
		clusterURLs  = []string{cfg.Elasticsearch.Url} // TODO: Add more URLs if necessary
		retryBackoff = backoff.NewExponentialBackOff()
	)

	esCfg := es7.Config{
		Addresses: clusterURLs,
		Username:  cfg.Elasticsearch.Username,
		Password:  cfg.Elasticsearch.Password,

		RetryOnStatus: []int{502, 503, 504, 429},

		// Configure the backoff function
		//
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},

		// Retry up to 5 attempts
		//
		MaxRetries: 5,
	}

	es7Client, _ := es7.NewClient(esCfg)

	fmt.Printf("Using ES7 Client Version %s\n", es7.Version)

	return es7Client
}
