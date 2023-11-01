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

	"gohan/api/models"
	c "gohan/api/models/constants"
	a "gohan/api/models/constants/assembly-id"
	gq "gohan/api/models/constants/genotype-query"
	s "gohan/api/models/constants/sort"
	z "gohan/api/models/constants/zygosity"
	"gohan/api/utils"

	"github.com/elastic/go-elasticsearch/v7"
)

const wildcardVariantsIndex = "variants-*"

func GetDocumentsByDocumentId(cfg *models.Config, es *elasticsearch.Client, id string) (map[string]interface{}, error) {

	// overall query structure
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{{
					"bool": map[string]interface{}{
						"must": []map[string]interface{}{
							{
								"query_string": map[string]string{
									"query": fmt.Sprintf("_id:%s", id),
								},
							},
						},
					}},
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

	fmt.Printf("Query Start: %s\n", time.Now())

	if cfg.Debug {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	// Perform the search request.
	res, searchErr := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(wildcardVariantsIndex),
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

	umErr := json.Unmarshal([]byte(jsonBodyString), &result)
	if umErr != nil {
		fmt.Printf("Error unmarshalling response: %s\n", umErr)
		return nil, umErr
	}

	fmt.Printf("Query End: %s\n", time.Now())

	return result, nil
}

func GetDocumentsContainerVariantOrSampleIdInPositionRange(cfg *models.Config, es *elasticsearch.Client,
	chromosome string, lowerBound int, upperBound int,
	variantId string, sampleId string, datasetString string,
	reference string, alternative string, alleles []string,
	size int, sortByPosition c.SortDirection,
	includeInfoInResultSet bool,
	genotype c.GenotypeQuery, assemblyId c.AssemblyId,
	getSampleIdsOnly bool) (map[string]interface{}, error) {

	// begin building the request body.
	mustMap := []map[string]interface{}{{
		"query_string": map[string]interface{}{
			"query": "chrom:" + chromosome,
		}},
	}
	var (
		shouldMap          []map[string]interface{}
		minimumShouldMatch int
	)

	// 'complexifying' the query
	// TODO: refactor common code between 'Get' and 'Count'-DocumentsContainerVariantOrSampleIdInPositionRange
	if variantId != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"query_string": map[string]interface{}{
				"fields": []string{"id"},
				"query":  variantId,
			},
		})
	}

	if sampleId != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"query_string": map[string]interface{}{
				"fields": []string{"sample.id"},
				"query":  sampleId,
			},
		})
	}

	if datasetString != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"query_string": map[string]interface{}{
				"fields": []string{"dataset.keyword"},
				"query":  datasetString,
			},
		})
	}

	if alternative != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": "alt:" + alternative,
			}})
	}

	shouldMap, minimumShouldMatch = addAllelesToShouldMap(alleles, genotype, shouldMap)

	if reference != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": "ref:" + reference,
			}})
	}

	if assemblyId != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"match": map[string]interface{}{
				"assemblyId": map[string]interface{}{
					"query": assemblyId,
				},
			},
		})
	}

	rangeMapSlice := []map[string]interface{}{}

	// TODO: make upperbound and lowerbound nilable, somehow?
	if upperBound > 0 {
		rangeMapSlice = append(rangeMapSlice, map[string]interface{}{
			"range": map[string]interface{}{
				"pos": map[string]interface{}{
					"lte": upperBound,
				},
			},
		})
	}

	if lowerBound > 0 {
		rangeMapSlice = append(rangeMapSlice, map[string]interface{}{
			"range": map[string]interface{}{
				"pos": map[string]interface{}{
					"gte": lowerBound,
				},
			},
		})
	}

	if genotype != gq.UNCALLED {
		mustMap = addZygosityToMustMap(genotype, mustMap)
	}

	// individually append each range components to the must map
	if len(rangeMapSlice) > 0 {
		for _, rms := range rangeMapSlice {
			mustMap = append(mustMap, rms)
		}
	}

	// exclude samples from result?
	var excludesSlice []string = make([]string, 0)
	if !includeInfoInResultSet {
		excludesSlice = append(excludesSlice, "info")
	}

	// overall query structure
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{{
					"bool": map[string]interface{}{
						"must":                 mustMap,
						"should":               shouldMap,
						"minimum_should_match": minimumShouldMatch,
					}},
				},
			},
		},
	}

	if !getSampleIdsOnly {
		query["size"] = size
		query["_source"] = map[string]interface{}{
			"includes": [1]string{"*"}, // include every field except those that may be specified in the 'excludesSlice'
			"excludes": excludesSlice,
		}
	} else {
		query["size"] = 0                       // return no full variant documents
		query["aggs"] = map[string]interface{}{ // aggregate only sample ids
			"sampleIds": map[string]interface{}{
				"terms": map[string]interface{}{
					"size":  "10000", // increases the number of buckets returned (default is 10)
					"field": "sample.id.keyword",
				},
			},
		}
	}

	// set up sorting
	if sortByPosition == s.Undefined {
		// default to ascending order
		sortByPosition = s.Ascending
	}

	// append sorting components
	query["sort"] = map[string]string{
		"pos": string(sortByPosition),
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

	fmt.Printf("Query Start: %s\n", time.Now())

	if cfg.Debug {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	// Perform the search request.
	res, searchErr := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(wildcardVariantsIndex),
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
		return nil, fmt.Errorf("failed to get variants by id : got '%s'", bracketString)
	}

	umErr := json.Unmarshal([]byte(jsonBodyString), &result)
	if umErr != nil {
		fmt.Printf("Error unmarshalling response: %s\n", umErr)
		return nil, umErr
	}

	fmt.Printf("Query End: %s\n", time.Now())

	return result, nil
}

func GetMostRecentVariantTimestamp(cfg *models.Config, es *elasticsearch.Client, dataset string) (time.Time, error) {
	// Initialize a zero-value timestamp
	var mostRecentTimestamp time.Time

	// Setup the Elasticsearch query to fetch the most recent 'created' timestamp
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]string{
				"dataset.keyword": dataset,
			},
		},
		"size": 1,
		"sort": []map[string]interface{}{
			{
				"createdTime": map[string]string{
					"order": "desc",
				},
			},
		},
	}

	// Encode the query to JSON
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s\n", err)
		return mostRecentTimestamp, err
	}

	// Print the constructed query for debugging
	fmt.Println("Constructed Elasticsearch Query:", string(buf.Bytes()))

	// Execute the query against Elasticsearch
	res, searchErr := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(wildcardVariantsIndex),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)

	if searchErr != nil {
		fmt.Printf("Error executing search request: %s\n", searchErr)
		return mostRecentTimestamp, searchErr
	}
	defer res.Body.Close()

	// Parse the response
	var result map[string]interface{}
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&result); err != nil {
		fmt.Printf("Error unmarshalling Elasticsearch response: %s\n", err)
		return mostRecentTimestamp, err
	}

	// Extract the 'created' timestamp from the first hit (if available)
	if hits, found := result["hits"].(map[string]interface{}); found {
		if hitSlice, hitFound := hits["hits"].([]interface{}); hitFound && len(hitSlice) > 0 {
			if firstHit, firstHitFound := hitSlice[0].(map[string]interface{}); firstHitFound {
				if source, sourceFound := firstHit["_source"].(map[string]interface{}); sourceFound {
					if created, createdFound := source["createdTime"].(string); createdFound {
						parsedTime, err := time.Parse(time.RFC3339, created)
						if err == nil {
							mostRecentTimestamp = parsedTime
						} else {
							fmt.Printf("Error parsing 'createdTime' timestamp: %s\n", err)
						}
					}
				}
			}
		} else {
			fmt.Println("No hits found for dataset:", dataset)
		}
	}

	return mostRecentTimestamp, nil
}

func CountDocumentsContainerVariantOrSampleIdInPositionRange(cfg *models.Config, es *elasticsearch.Client,
	chromosome string, lowerBound int, upperBound int,
	variantId string, sampleId string, datasetString string,
	reference string, alternative string, alleles []string,
	genotype c.GenotypeQuery, assemblyId c.AssemblyId) (map[string]interface{}, error) {

	// begin building the request body.
	mustMap := []map[string]interface{}{{
		"query_string": map[string]interface{}{
			"query": "chrom:" + chromosome,
		}},
	}
	var (
		shouldMap          []map[string]interface{}
		minimumShouldMatch int
	)

	// 'complexifying' the query
	// TODO: refactor common code between 'Get' and 'Count'-DocumentsContainerVariantOrSampleIdInPositionRange
	matchMap := make(map[string]interface{})

	if variantId != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"query_string": map[string]interface{}{
				"fields": []string{"id"},
				"query":  variantId,
			},
		})
	}

	if sampleId != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"query_string": map[string]interface{}{
				"fields": []string{"sample.id"},
				"query":  sampleId,
			},
		})
	}

	if datasetString != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"query_string": map[string]interface{}{
				"fields": []string{"dataset.keyword"},
				"query":  datasetString,
			},
		})
	}

	if alternative != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": "alt:" + alternative,
			}})
	}

	shouldMap, minimumShouldMatch = addAllelesToShouldMap(alleles, genotype, shouldMap)

	if reference != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": "ref:" + reference,
			}})
	}

	if assemblyId != "" && assemblyId != a.Unknown {
		mustMap = append(mustMap, map[string]interface{}{
			"match": map[string]interface{}{
				"assemblyId": map[string]interface{}{
					"query": assemblyId,
				},
			},
		})
	}

	rangeMapSlice := []map[string]interface{}{}

	// TODO: make upperbound and lowerbound nilable, somehow?
	if upperBound > 0 {
		rangeMapSlice = append(rangeMapSlice, map[string]interface{}{
			"range": map[string]interface{}{
				"pos": map[string]interface{}{
					"lte": upperBound,
				},
			},
		})
	}

	if lowerBound > 0 {
		rangeMapSlice = append(rangeMapSlice, map[string]interface{}{
			"range": map[string]interface{}{
				"pos": map[string]interface{}{
					"gte": lowerBound,
				},
			},
		})
	}

	if genotype != gq.UNCALLED {
		mustMap = addZygosityToMustMap(genotype, mustMap)
	}

	// append the match components to the must map
	if len(matchMap) > 0 {
		mustMap = append(mustMap, map[string]interface{}{
			"match": matchMap,
		})
	}

	// individually append each range components to the must map
	if len(rangeMapSlice) > 0 {
		for _, rms := range rangeMapSlice {
			mustMap = append(mustMap, rms)
		}
	}

	// overall query structure
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{{
					"bool": map[string]interface{}{
						"must":                 mustMap,
						"should":               shouldMap,
						"minimum_should_match": minimumShouldMatch,
					}},
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

	fmt.Printf("Query Start: %s\n", time.Now())

	if cfg.Debug {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	// Perform the search request.
	res, searchErr := es.Count(
		es.Count.WithContext(context.Background()),
		es.Count.WithIndex(wildcardVariantsIndex),
		es.Count.WithBody(&buf),
		es.Count.WithPretty(),
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
		return nil, fmt.Errorf("failed to count variants by id : got '%s'", bracketString)
	}

	umErr := json.Unmarshal([]byte(jsonBodyString), &result)
	if umErr != nil {
		fmt.Printf("Error unmarshalling response: %s\n", umErr)
		return nil, umErr
	}

	fmt.Printf("Query End: %s\n", time.Now())

	return result, nil
}

func GetVariantsBucketsByKeyword(cfg *models.Config, es *elasticsearch.Client, keyword string) (map[string]interface{}, error) {
	// begin building the request body.
	var buf bytes.Buffer
	aggMap := map[string]interface{}{
		"size": "0",
		"aggs": map[string]interface{}{
			"items": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": keyword,
					"size":  "10000", // increases the number of buckets returned (default is 10)
					"order": map[string]string{
						"_key": "asc",
					},
				},
			},
			"latest_created": map[string]interface{}{
				"max": map[string]interface{}{
					"field": "createdTime",
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
	}

	if cfg.Debug {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	// Perform the search request.
	res, searchErr := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(wildcardVariantsIndex),
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
		return nil, fmt.Errorf("failed to get buckets by keyword: got '%s'", bracketString)
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

func GetVariantsBucketsByKeywordAndDataset(cfg *models.Config, es *elasticsearch.Client, keyword string, dataset string) (map[string]interface{}, error) {
	// begin building the request body.
	var buf bytes.Buffer
	aggMap := map[string]interface{}{
		"size": "0",
		"aggs": map[string]interface{}{
			"items": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": keyword,
					"size":  "10000", // increases the number of buckets returned (default is 10)
					"order": map[string]string{
						"_key": "asc",
					},
				},
			},
		},
	}

	if dataset != "" {
		aggMap["query"] = map[string]interface{}{
			"match": map[string]interface{}{
				"dataset": dataset,
			},
		}
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
	}

	fmt.Printf("Query Start: %s\n", time.Now())

	if cfg.Debug {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	// Perform the search request.
	res, searchErr := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(wildcardVariantsIndex),
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
		return nil, fmt.Errorf("failed to get buckets by keyword: got '%s'", bracketString)
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

func DeleteVariantsByDatasetId(cfg *models.Config, es *elasticsearch.Client, dataset string) (map[string]interface{}, error) {

	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"dataset": dataset,
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s\n", query)
	}

	if cfg.Debug {
		// view the outbound elasticsearch query
		myString := string(buf.Bytes()[:])
		fmt.Println(myString)
	}

	// Perform the delete request.
	deleteRes, deleteErr := es.DeleteByQuery(
		[]string{wildcardVariantsIndex},
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
	umErr := json.Unmarshal([]byte(jsonBodyString), &result)
	if umErr != nil {
		fmt.Printf("Error unmarshalling variant deletion response: %s\n", umErr)
		return nil, umErr
	}

	return result, nil
}

// -- internal use only --
func addAllelesToShouldMap(alleles []string, genotype c.GenotypeQuery, allelesShouldMap []map[string]interface{}) ([]map[string]interface{}, int) {
	minimumShouldMatch := 0

	if len(alleles) > 0 {
		switch len(alleles) {
		case 1:
			if genotype == gq.ALTERNATE || genotype == gq.REFERENCE {
				// haploid case
				// - queried allele should be present on the left side of the pair with an empty right side
				allelesShouldMap = append(allelesShouldMap, allelesShouldMapBuilder(alleles[0], "AND", "\"\""))

			} else {
				// assume diploid-type of search as default
				// - queried allele can be present on either side of the pair
				allelesShouldMap = append(allelesShouldMap, allelesShouldMapBuilder(alleles[0], "OR", alleles[0]))
			}
		case 2:
			if genotype == gq.ALTERNATE || genotype == gq.REFERENCE {
				// haploid case
				// - either queried allele can be present on the left side of the pair with an empty right side
				allelesShouldMap = append(allelesShouldMap, allelesShouldMapBuilder(alleles[0], "AND", "\"\""))
				allelesShouldMap = append(allelesShouldMap, allelesShouldMapBuilder(alleles[1], "AND", "\"\""))

			} else {
				// assume diploid-type of search as default
				// - treat as a left/right pair;
				//   either queried allele can be present on the left or right side of the pair
				allelesShouldMap = append(allelesShouldMap, allelesShouldMapBuilder(alleles[0], "AND", alleles[1]))
				allelesShouldMap = append(allelesShouldMap, allelesShouldMapBuilder(alleles[1], "AND", alleles[0]))
			}
			// TODO: triploid ?
		}
		minimumShouldMatch = 1
	}

	return allelesShouldMap, minimumShouldMatch
}
func allelesShouldMapBuilder(alleleLeft string, operator string, alleleRight string) map[string]interface{} {
	return map[string]interface{}{
		"query_string": map[string]interface{}{
			"query": "sample.variation.alleles.left.keyword:" + alleleLeft + " " + operator + " sample.variation.alleles.right.keyword:" + alleleRight,
		}}
}

func addZygosityToMustMap(genotype c.GenotypeQuery, mustMap []map[string]interface{}) []map[string]interface{} {
	zygosityMatchMap := make(map[string]interface{})

	switch genotype {
	// Haploid
	case gq.REFERENCE:
		zygosityMatchMap["sample.variation.genotype.zygosity"] = map[string]interface{}{
			"query": z.Reference,
		}

	case gq.ALTERNATE:
		zygosityMatchMap["sample.variation.genotype.zygosity"] = map[string]interface{}{
			"query": z.Alternate,
		}
	// Diploid
	case gq.HETEROZYGOUS:
		zygosityMatchMap["sample.variation.genotype.zygosity"] = map[string]interface{}{
			"query": z.Heterozygous,
		}

	case gq.HOMOZYGOUS_REFERENCE:
		zygosityMatchMap["sample.variation.genotype.zygosity"] = map[string]interface{}{
			"query": z.HomozygousReference,
		}

	case gq.HOMOZYGOUS_ALTERNATE:
		zygosityMatchMap["sample.variation.genotype.zygosity"] = map[string]interface{}{
			"query": z.HomozygousAlternate,
		}
	}

	// - verify zygosity-map is not empty before adding to the must-map
	if len(zygosityMatchMap) > 0 {
		mustMap = append(mustMap, map[string]interface{}{
			"match": zygosityMatchMap,
		})
	}

	return mustMap
}
