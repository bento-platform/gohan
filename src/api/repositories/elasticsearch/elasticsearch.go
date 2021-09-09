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

	c "api/models/constants"
	gq "api/models/constants/genotype-query"
	z "api/models/constants/zygosity"

	"github.com/elastic/go-elasticsearch"
)

func GetDocumentsContainerVariantOrSampleIdInPositionRange(es *elasticsearch.Client,
	chromosome string, lowerBound int, upperBound int,
	variantId string, sampleId string,
	reference string, alternative string,
	size int, sortByPosition string,
	includeSamplesInResultSet bool, genotype c.GenotypeQuery) map[string]interface{} {

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
		matchMap["id"] = map[string]interface{}{
			"query": variantId,
		}
	}

	if sampleId != "" {
		matchMap["samples.id"] = map[string]interface{}{
			"query": sampleId,
		}
	}

	if alternative != "" {
		matchMap["alt"] = map[string]interface{}{
			"query": alternative,
		}
	}

	if reference != "" {
		matchMap["ref"] = map[string]interface{}{
			"query": reference,
		}
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

	// exclude samples from result?
	var excludesSlice []string = make([]string, 0)
	if includeSamplesInResultSet == false {
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
	sortDirection := "asc"
	if sortByPosition != "" {
		switch sortByPosition {
		case "asc":
			fmt.Println("Already set 'sortByPosition' keyword 'asc' to query")
			break
		case "desc":
			fmt.Println("Setting 'sortByPosition' keyword 'desc' to query")
			sortDirection = "desc"
			break
		default:
			fmt.Printf("Found unknown 'sortByPosition' keyword : %s -- ignoring\n", sortByPosition)
			break
		}

	} else {
		fmt.Println("Found empty 'sortByPosition' keyword -- defaulting to 'asc'")
	}

	// append sorting components
	query["sort"] = map[string]string{
		"pos": sortDirection,
	}

	// encode the query
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s\n", err)
	}

	// DEBUG--
	// Unmarshal or Decode the JSON to the interface.
	myString := string(buf.Bytes()[:])
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
	resultString := res.String()
	fmt.Println(resultString)
	// --

	// Declared an empty interface
	result := make(map[string]interface{})

	// Unmarshal or Decode the JSON to the interface.
	// Known bug: response comes back with a preceing '[200 OK] ' which needs trimming (hence the [9:])
	umErr := json.Unmarshal([]byte(resultString[9:]), &result)
	if umErr != nil {
		fmt.Printf("Error unmarshalling response: %s\n", umErr)
	}

	fmt.Printf("Query End: %s\n", time.Now())

	return result
}

func CountDocumentsContainerVariantOrSampleIdInPositionRange(es *elasticsearch.Client,
	chromosome string, lowerBound int, upperBound int,
	variantId string, sampleId string,
	reference string, alternative string, genotype c.GenotypeQuery) map[string]interface{} {

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
		matchMap["id"] = map[string]interface{}{
			"query": variantId,
		}
	}

	if sampleId != "" {
		matchMap["samples.id"] = map[string]interface{}{
			"query": sampleId,
		}
	}

	if alternative != "" {
		matchMap["alt"] = map[string]interface{}{
			"query": alternative,
		}
	}

	if reference != "" {
		matchMap["ref"] = map[string]interface{}{
			"query": reference,
		}
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

	// exclude samples from result?

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

	// DEBUG--
	// Unmarshal or Decode the JSON to the interface.
	myString := string(buf.Bytes()[:])
	fmt.Println(myString)
	// --

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

	// Temp
	resultString := res.String()
	// fmt.Println(resultString)
	// --

	// Declared an empty interface
	result := make(map[string]interface{})

	// Unmarshal or Decode the JSON to the interface.
	// Known bug: response comes back with a preceing '[200 OK] ' which needs trimming (hence the [9:])
	umErr := json.Unmarshal([]byte(resultString[9:]), &result)
	if umErr != nil {
		fmt.Printf("Error unmarshalling response: %s\n", umErr)
	}

	fmt.Printf("Query End: %s\n", time.Now())

	return result
}

func GetBucketsByKeyword(es *elasticsearch.Client, keyword string) map[string]interface{} {

	// begin building the request body.
	var buf bytes.Buffer
	aggMap := map[string]interface{}{
		"size": "0",
		"aggs": map[string]interface{}{
			"items": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": keyword,
				},
			},
		},
	}

	// encode the query
	if err := json.NewEncoder(&buf).Encode(aggMap); err != nil {
		log.Fatalf("Error encoding aggMap: %s\n", err)
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

	// Temp
	resultString := res.String()
	//fmt.Println(resultString)
	// --

	// Declared an empty interface
	result := make(map[string]interface{})

	// Unmarshal or Decode the JSON to the interface.
	// Known bug: response comes back with a preceing '[200 OK] ' which needs trimming (hence the [9:])
	umErr := json.Unmarshal([]byte(resultString[9:]), &result)
	if umErr != nil {
		fmt.Printf("Error unmarshalling response: %s\n", umErr)
	}

	fmt.Printf("Query End: %s\n", time.Now())

	return result
}
