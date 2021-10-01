package elasticsearch

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"api/models"
	"api/models/constants"
	assemblyId "api/models/constants/assembly-id"

	"github.com/elastic/go-elasticsearch"
)

const genesIndex = "genes"

func GetGeneBucketsByKeyword(cfg *models.Config, es *elasticsearch.Client) map[string]interface{} {
	// begin building the request body.
	var buf bytes.Buffer
	aggMap := map[string]interface{}{
		"size": "0",
		"aggs": map[string]interface{}{
			"genes_assembly_id_group": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "assemblyId.keyword",
					"size":  "10000", // increases the number of buckets returned (default is 10)
					"order": map[string]string{
						"_key": "asc",
					},
				},
				"aggs": map[string]interface{}{
					"genes_chromosome_group": map[string]interface{}{
						"terms": map[string]interface{}{
							"field": "chrom.keyword",
							"size":  "10000", // increases the number of buckets returned (default is 10)
							"order": map[string]string{
								"_key": "asc",
							},
						},
					},
				},
			},
		},
	}

	// encode the query
	if err := json.NewEncoder(&buf).Encode(aggMap); err != nil {
		log.Fatalf("Error encoding aggMap: %s\n", err)
	}

	if cfg.Debug {
		// view the outbound elasticsearch query
		myString := string(buf.Bytes()[:])
		fmt.Println(myString)
	}

	// TEMP: SECURITY RISK
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//
	// Perform the search request.
	res, searchErr := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(genesIndex),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if searchErr != nil {
		fmt.Printf("Error getting response: %s\n", searchErr)
	}

	defer res.Body.Close()

	resultString := res.String()
	if cfg.Debug {
		fmt.Println(resultString)
	}

	// Declared an empty interface
	result := make(map[string]interface{})

	// Unmarshal or Decode the JSON to the interface.
	// Known bug: response comes back with a preceding '[200 OK] ' which needs trimming (hence the [9:])
	umErr := json.Unmarshal([]byte(resultString[9:]), &result)
	if umErr != nil {
		fmt.Printf("Error unmarshalling response: %s\n", umErr)
	}

	fmt.Printf("Query End: %s\n", time.Now())

	return result
}

func GetGeneDocumentsByTermWildcard(cfg *models.Config, es *elasticsearch.Client,
	chromosomeSearchTerm string, term string, assId constants.AssemblyId, size int) map[string]interface{} {

	// TEMP: SECURITY RISK
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//

	// Nomenclature Search Term
	nomenclatureStringTerm := fmt.Sprintf("*%s*", term)

	// Assembly Id Search Term (wildcard by default)
	assemblyIdStringTerm := "*"
	if assId != assemblyId.Unknown {
		assemblyIdStringTerm = string(assId)
	}

	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{{
					"bool": map[string]interface{}{
						"must": []map[string]interface{}{
							{
								"query_string": map[string]interface{}{
									"fields": []string{"chrom"},
									"query":  chromosomeSearchTerm,
								},
							},
							{
								"query_string": map[string]interface{}{
									"fields": []string{"name"},
									"query":  nomenclatureStringTerm,
								},
							},
							{
								"query_string": map[string]interface{}{
									"fields": []string{"assemblyId"},
									"query":  assemblyIdStringTerm,
								},
							},
						},
					},
				}},
			},
		},
		"size": size,
		"sort": []map[string]interface{}{
			{
				"chrom.keyword": map[string]interface{}{
					"order": "asc",
				},
			},
			{
				"start": map[string]interface{}{
					"order": "asc",
				},
			},
		},
	}

	// encode the query
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s\n", err)
	}

	if cfg.Debug {
		// view the outbound elasticsearch query
		myString := string(buf.Bytes()[:])
		fmt.Println(myString)
	}

	// Perform the search request.
	searchRes, searchErr := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(genesIndex),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if searchErr != nil {
		fmt.Printf("Error getting response: %s\n", searchErr)
	}

	defer searchRes.Body.Close()

	resultString := searchRes.String()
	if cfg.Debug {
		fmt.Println(resultString)
	}

	// Prepare an empty interface
	result := make(map[string]interface{})

	// Unmarshal or Decode the JSON to the empty interface.
	// Known bug: response comes back with a preceding '[200 OK] ' which needs trimming (hence the [9:])
	umErr := json.Unmarshal([]byte(resultString[9:]), &result)
	if umErr != nil {
		fmt.Printf("Error unmarshalling gene search response: %s\n", umErr)
	}

	return result
}
