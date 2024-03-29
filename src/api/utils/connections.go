package utils

import (
	"crypto/tls"
	"gohan/api/models"
	"net"
	"net/http"

	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	es7 "github.com/elastic/go-elasticsearch/v7"
)

func CreateEsConnection(cfg *models.Config) *es7.Client {
	var (
		clusterURLs  = []string{cfg.Elasticsearch.Url} // TODO: Add more URLs if necessary
		retryBackoff = backoff.NewExponentialBackOff()
	)

	// TODO: configure 'tr' based on debug status
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: time.Duration(2 / time.Second),
		}).DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	esCfg := es7.Config{
		Addresses: clusterURLs,
		Username:  cfg.Elasticsearch.Username,
		Password:  cfg.Elasticsearch.Password,

		RetryOnStatus: []int{502, 503, 504, 429},

		Transport: tr,

		// Configure the backoff function
		//
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},

		// Retry up to 2 attempts
		//
		MaxRetries:           2,
		EnableRetryOnTimeout: false,
	}

	es7Client, _ := es7.NewClient(esCfg)

	fmt.Printf("Using ES7 Client Version %s\n", es7.Version)

	return es7Client
}
