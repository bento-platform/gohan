package elasticsearch

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"api/models"
	"api/models/constants"
	assemblyId "api/models/constants/assembly-id"
	"api/utils"

	"github.com/elastic/go-elasticsearch/v7"
)

const genesIndex = "genes"

func GetGeneBucketsByKeyword(cfg *models.Config, es *elasticsearch.Client) (map[string]interface{}, error) {
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
		return nil, err
	}

	if cfg.Debug {
		// view the outbound elasticsearch query
		myString := string(buf.Bytes()[:])
		fmt.Println(myString)

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

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
		return nil, searchErr
	}

	defer res.Body.Close()

	resultString := res.String()
	if cfg.Debug {
		fmt.Println(resultString)
	}

	// Declared an empty interface
	result := make(map[string]interface{})

	// Unmarshal or Decode the JSON to the interface.
	// Known bug: response comes back with a preceding '[200 OK] ' which needs trimming
	bracketString, jsonBodyString := utils.GetLeadingStringInBetweenSquareBrackets(resultString)
	if !strings.Contains(bracketString, "200") {
		return nil, fmt.Errorf("failed to get documents by id : got '%s'", bracketString)
	}
	// umErr := json.Unmarshal([]byte(resultString[9:]), &result)
	umErr := json.Unmarshal([]byte(jsonBodyString), &result)
	if umErr != nil {
		fmt.Printf("Error unmarshalling response: %s\n", umErr)
		return nil, umErr
	}

	fmt.Printf("Query End: %s\n", time.Now())

	return result, nil
}

func GetGeneDocumentsByTermWildcard(cfg *models.Config, es *elasticsearch.Client,
	chromosomeSearchTerm string, term string, assId constants.AssemblyId, size int) (map[string]interface{}, error) {

	if cfg.Debug {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

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
		return nil, err
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
		return nil, searchErr
	}

	defer searchRes.Body.Close()

	resultString := searchRes.String()
	if cfg.Debug {
		fmt.Println(resultString)
	}

	// Prepare an empty interface
	result := make(map[string]interface{})

	// Unmarshal or Decode the JSON to the empty interface.
	// Known bug: response comes back with a preceding '[200 OK] ' which needs trimming
	bracketString, jsonBodyString := utils.GetLeadingStringInBetweenSquareBrackets(resultString)
	if !strings.Contains(bracketString, "200") {
		return nil, fmt.Errorf("failed to get documents by id : got '%s'", bracketString)
	}
	// umErr := json.Unmarshal([]byte(resultString[9:]), &result)
	umErr := json.Unmarshal([]byte(jsonBodyString), &result)
	if umErr != nil {
		fmt.Printf("Error unmarshalling gene search response: %s\n", umErr)
		return nil, umErr
	}

	return result, nil
}

func DeleteGenesByAssemblyId(cfg *models.Config, es *elasticsearch.Client, assId constants.AssemblyId) (map[string]interface{}, error) {

	if cfg.Debug {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"assemblyId": string(assId),
			},
		},
	}

	// encode the query
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s\n", err)
		return nil, err
	}

	if cfg.Debug {
		// view the outbound elasticsearch query
		myString := string(buf.Bytes()[:])
		fmt.Println(myString)
	}

	// Perform the delete request.
	deleteRes, deleteErr := es.DeleteByQuery(
		[]string{genesIndex},
		bytes.NewReader(buf.Bytes()),
	)
	if deleteErr != nil {
		fmt.Printf("Error getting response: %s\n", deleteErr)
		return nil, deleteErr
	}

	defer deleteRes.Body.Close()

	resultString := deleteRes.String()
	if cfg.Debug {
		fmt.Println(resultString)
	}

	// Prepare an empty interface
	result := make(map[string]interface{})

	// Unmarshal or Decode the JSON to the empty interface.
	// Known bug: response comes back with a preceding '[200 OK] ' which needs trimming
	bracketString, jsonBodyString := utils.GetLeadingStringInBetweenSquareBrackets(resultString)
	if !strings.Contains(bracketString, "200") {
		return nil, fmt.Errorf("failed to get documents by id : got '%s'", bracketString)
	}
	// umErr := json.Unmarshal([]byte(resultString[9:]), &result)
	umErr := json.Unmarshal([]byte(jsonBodyString), &result)
	if umErr != nil {
		fmt.Printf("Error unmarshalling gene search response: %s\n", umErr)
		return nil, umErr
	}

	return result, nil
}
