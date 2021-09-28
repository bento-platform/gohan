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
	c "api/models/constants"
	gq "api/models/constants/genotype-query"
	s "api/models/constants/sort"
	z "api/models/constants/zygosity"

	"github.com/elastic/go-elasticsearch"
)

func GetDocumentsContainerVariantOrSampleIdInPositionRange(cfg *models.Config, es *elasticsearch.Client,
	chromosome string, lowerBound int, upperBound int,
	variantId string, sampleId string,
	reference string, alternative string,
	size int, sortByPosition c.SortDirection,
	includeSamplesInResultSet bool,
	genotype c.GenotypeQuery, assemblyId c.AssemblyId) map[string]interface{} {

	// begin building the request body.
	mustMap := []map[string]interface{}{{
		"query_string": map[string]interface{}{
			"query": "chrom:" + chromosome,
		}},
	}

	// 'complexifying' the query
	// TODO: refactor common code between 'Get' and 'Count'-DocumentsContainerVariantOrSampleIdInPositionRange
	if variantId != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"match": map[string]interface{}{
				"id": map[string]interface{}{
					"query": variantId,
				},
			},
		})

	}

	if sampleId != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"match": map[string]interface{}{
				"samples.id": map[string]interface{}{
					"query": sampleId,
				},
			},
		})
	}

	if alternative != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"match": map[string]interface{}{
				"alt": map[string]interface{}{
					"query": alternative,
				},
			},
		})
	}

	if reference != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"match": map[string]interface{}{
				"ref": map[string]interface{}{
					"query": reference,
				},
			},
		})
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
		zygosityMatchMap := make(map[string]interface{})

		switch genotype {
		case gq.HETEROZYGOUS:
			zygosityMatchMap["samples.variation.genotype.zygosity"] = map[string]interface{}{
				"query": z.Heterozygous,
			}

		case gq.HOMOZYGOUS_REFERENCE:
			zygosityMatchMap["samples.variation.genotype.zygosity"] = map[string]interface{}{
				"query": z.Homozygous,
			}

			mustMap = append(mustMap, map[string]interface{}{
				"match": map[string]interface{}{
					"samples.variation.genotype.alleleLeft": map[string]interface{}{
						"query": 0,
					},
				},
			})

		case gq.HOMOZYGOUS_ALTERNATE:
			zygosityMatchMap["samples.variation.genotype.zygosity"] = map[string]interface{}{
				"query": z.Homozygous,
			}

			rangeMapSlice = append(rangeMapSlice, map[string]interface{}{
				"range": map[string]interface{}{
					"samples.variation.genotype.alleleLeft": map[string]interface{}{
						"gte": 0,
					},
				},
			})
		}

		mustMap = append(mustMap, map[string]interface{}{
			"match": zygosityMatchMap,
		})
	}

	// individually append each range components to the must map
	if len(rangeMapSlice) > 0 {
		for _, rms := range rangeMapSlice {
			mustMap = append(mustMap, rms)
		}
	}

	// exclude samples from result?
	var excludesSlice []string = make([]string, 0)
	if !includeSamplesInResultSet {
		excludesSlice = append(excludesSlice, "samples")
	}

	// overall query structure
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{{
					"bool": map[string]interface{}{
						"must": mustMap,
					}},
				},
			},
		},
		"_source": map[string]interface{}{
			"includes": [1]string{"*"}, // include every field except those that may be specified in the 'excludesSlice'
			"excludes": excludesSlice,
		},
		"size": size,
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
	}

	if cfg.Debug {
		// view the outbound elasticsearch query
		myString := string(buf.Bytes()[:])
		fmt.Println(myString)
	}

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

func CountDocumentsContainerVariantOrSampleIdInPositionRange(cfg *models.Config, es *elasticsearch.Client,
	chromosome string, lowerBound int, upperBound int,
	variantId string, sampleId string,
	reference string, alternative string,
	genotype c.GenotypeQuery, assemblyId c.AssemblyId) map[string]interface{} {

	// begin building the request body.
	mustMap := []map[string]interface{}{{
		"query_string": map[string]interface{}{
			"query": "chrom:" + chromosome,
		}},
	}

	// 'complexifying' the query
	// TODO: refactor common code between 'Get' and 'Count'-DocumentsContainerVariantOrSampleIdInPositionRange
	matchMap := make(map[string]interface{})

	if variantId != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"match": map[string]interface{}{
				"id": map[string]interface{}{
					"query": variantId,
				},
			},
		})

	}

	if sampleId != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"match": map[string]interface{}{
				"samples.id": map[string]interface{}{
					"query": sampleId,
				},
			},
		})
	}

	if alternative != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"match": map[string]interface{}{
				"alt": map[string]interface{}{
					"query": alternative,
				},
			},
		})
	}

	if reference != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"match": map[string]interface{}{
				"ref": map[string]interface{}{
					"query": reference,
				},
			},
		})
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
		zygosityMatchMap := make(map[string]interface{})

		switch genotype {
		case gq.HETEROZYGOUS:
			zygosityMatchMap["samples.variation.genotype.zygosity"] = map[string]interface{}{
				"query": z.Heterozygous,
			}

		case gq.HOMOZYGOUS_REFERENCE:
			zygosityMatchMap["samples.variation.genotype.zygosity"] = map[string]interface{}{
				"query": z.Homozygous,
			}

			mustMap = append(mustMap, map[string]interface{}{
				"match": map[string]interface{}{
					"samples.variation.genotype.alleleLeft": map[string]interface{}{
						"query": 0,
					},
				},
			})

		case gq.HOMOZYGOUS_ALTERNATE:
			zygosityMatchMap["samples.variation.genotype.zygosity"] = map[string]interface{}{
				"query": z.Homozygous,
			}

			rangeMapSlice = append(rangeMapSlice, map[string]interface{}{
				"range": map[string]interface{}{
					"samples.variation.genotype.alleleLeft": map[string]interface{}{
						"gte": 0,
					},
				},
			})
		}

		mustMap = append(mustMap, map[string]interface{}{
			"match": zygosityMatchMap,
		})
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
						"must": mustMap,
					}},
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

	fmt.Printf("Query Start: %s\n", time.Now())

	// TEMP: SECURITY RISK
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//
	// Perform the search request.
	res, searchErr := es.Count(
		es.Count.WithContext(context.Background()),
		es.Count.WithIndex("variants"),
		es.Count.WithBody(&buf),
		es.Count.WithPretty(),
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

func GetBucketsByKeyword(cfg *models.Config, es *elasticsearch.Client, keyword string) map[string]interface{} {

	// begin building the request body.
	var buf bytes.Buffer
	aggMap := map[string]interface{}{
		"size": "0",
		"aggs": map[string]interface{}{
			"items": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": keyword,
					"size":  "10000", // increases the number of buckets returned (default is 10)
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
		es.Search.WithIndex("variants"),
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

func GetGeneDocumentsByTermWildcard(cfg *models.Config, es *elasticsearch.Client, term string) map[string]interface{} {

	// TEMP: SECURITY RISK
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//

	// overall query structure
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"wildcard": map[string]interface{}{
				"nomenclature": map[string]interface{}{
					"value":   fmt.Sprintf("*%s*", term),
					"boost":   1.0,
					"rewrite": "constant_score",
				},
			},
		},
		"size": 25, // default
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
		es.Search.WithIndex("genes"),
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
