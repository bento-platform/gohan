package utils

import (
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/elastic/go-elasticsearch"
	es7 "github.com/elastic/go-elasticsearch"
)

func CreateEsConnection(elasticsearchUrl string, elasticsearchUsername string, elasticsearchPassword string) *es7.Client {
	var (
		clusterURLs  = []string{elasticsearchUrl} // TODO: Add more URLs if necessary
		retryBackoff = backoff.NewExponentialBackOff()
	)

	cfg := elasticsearch.Config{
		Addresses: clusterURLs,
		Username:  elasticsearchUsername,
		Password:  elasticsearchPassword,

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

	es7Client, _ := es7.NewClient(cfg)

	fmt.Printf("Using ES7 Client Version %s", es7.Version)

	return es7Client
}
