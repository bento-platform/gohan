package utils

import (
	"fmt"

	"github.com/elastic/go-elasticsearch"
	es7 "github.com/elastic/go-elasticsearch"
)

func CreateEsConnection(elasticsearchUrl string, elasticsearchUsername string, elasticsearchPassword string) *es7.Client {
	var clusterURLs = []string{elasticsearchUrl} // TODO: Add more URLs if necessary

	cfg := elasticsearch.Config{
		Addresses: clusterURLs,
		Username:  elasticsearchUsername,
		Password:  elasticsearchPassword,
	}
	es7Client, _ := es7.NewClient(cfg)

	fmt.Printf("Using ES7 Client Version %s", es7.Version)

	return es7Client
}
