package utils

import (
	"fmt"

	"github.com/elastic/go-elasticsearch"
	es7 "github.com/elastic/go-elasticsearch"
)

func CreateEsConnection() *es7.Client {
	var (
		// Testing
		clusterURLs = []string{"http://localhost:9200"}
		username    = "elastic"
		password    = "changeme!"
	)
	cfg := elasticsearch.Config{
		Addresses: clusterURLs,
		Username:  username,
		Password:  password,
	}
	es7Client, _ := es7.NewClient(cfg)

	fmt.Printf("Using ES7 Client Version %s", es7.Version)

	return es7Client
}
