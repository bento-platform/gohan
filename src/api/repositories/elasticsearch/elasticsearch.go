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

	"github.com/elastic/go-elasticsearch"
)

func GetDocumentsContainerVariantOrSampleIdInPositionRange(es *elasticsearch.Client,
	chromosome string, lowerBound int, upperBound int,
	variantId string, sampleId string,
	reference string, alternative string,
	size int, sortByPosition string,
	includeSamplesInResultSet bool) map[string]interface{} {
	// TODO : implement query

	// Build the request body.
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{{
					"query_string": map[string]interface{}{
						"query": "chrom:" + chromosome,
					}},
				},
			},
		},
	}

	// testing
	if sortByPosition != "" {
		sortDirection := ""

		switch sortByPosition {
		case "asc":
			fmt.Println("Appending 'sortByPosition' keyword 'asc' to query")
			sortDirection = "asc"
			break
		case "desc":
			fmt.Println("Appending 'sortByPosition' keyword 'desc' to query")
			sortDirection = "desc"
			break
		default:
			fmt.Printf("Found unknown 'sortByPosition' keyword : %s -- ignoring\n", sortByPosition)
			break
		}

		if sortDirection != "" {
			query["sort"] = map[string]string{
				"pos": sortDirection,
			}
		}
	}
	//

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	// DEBUG

	// Unmarshal or Decode the JSON to the interface.
	myString := string(buf.Bytes()[:])

	// Temp
	fmt.Println(myString)
	// --

	fmt.Printf("Query Start: %s\n", time.Now())

	// TEMP: SECURITY RISK
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//
	// Perform the search request.
	res, searchErr := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("variants"),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if searchErr != nil {
		fmt.Printf("Error getting response: %s\n", searchErr)
	}

	defer res.Body.Close()

	// Temp
	fmt.Println(res.String())
	resultString := res.String()
	// --

	// Declared an empty interface
	result := make(map[string]interface{})

	// Unmarshal or Decode the JSON to the interface.
	// Known bug: response comes back with a preceing [200 Success] which needs trimming (hence the [9:])
	umErr := json.Unmarshal([]byte(resultString[9:]), &result)
	if umErr != nil {
		fmt.Printf("Error unmarshalling response: %s\n", umErr)
	}

	fmt.Printf("Query End: %s", time.Now())

	return result
}
