package utils

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"gohan/api/models"
	c "gohan/api/models/constants/chromosome"
	"log"
	"net"
	"net/http"
	"strings"

	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/elastic/go-elasticsearch/esapi"
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
	createIndicesIfNecessary(es7Client)

	fmt.Printf("Using ES7 Client Version %s\n", es7.Version)

	return es7Client
}

func createIndicesIfNecessary(es7Client *es7.Client) {
	var indexModels = []string{"genes"}
	for _, humChrom := range c.ValidListOfHumanChromosomes() {
		indexModels = append(indexModels, strings.ToLower(fmt.Sprintf("variants-%s", humChrom)))
	}

	for _, imkey := range indexModels {
		// create indexes if not already existing
		// - check if it exists
		indicesExistsResp, err := esapi.IndicesExistsRequest{
			Index: []string{imkey},
		}.Do(context.Background(), es7Client)
		if err != nil {
			log.Fatal(err)
		}

		if indicesExistsResp.StatusCode != 200 {
			// - create index
			// -- mapping struct to elasticsearch
			var mappingStruct = make(map[string]interface{})
			// ignore specifying "mapping" for now. that will be
			// inferred at ingestion time

			// -- specify settings
			var nshards = 1 // default (for genes)
			if strings.Contains(imkey, "variants") {
				nshards = 2 // for each variants-* indexes
			}
			mappingStruct["settings"] = map[string]interface{}{
				"number_of_shards": nshards,
			}

			// -- push to es
			mappingJson, _ := json.Marshal(mappingStruct)
			res, err := es7Client.Indices.Create(
				imkey,
				es7Client.Indices.Create.WithBody(strings.NewReader(string(mappingJson))),
			)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Created index: %s\n", res)
		}
	}
}
