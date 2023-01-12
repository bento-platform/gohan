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
	es7 "github.com/elastic/go-elasticsearch/v7"
)

const variantsIndex = "variants"

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
		es.Search.WithIndex(variantsIndex),
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
	variantId string, sampleId string,
	reference string, alternative string, alleles []string,
	size int, sortByPosition c.SortDirection,
	includeInfoInResultSet bool,
	genotype c.GenotypeQuery, assemblyId c.AssemblyId, tableId string,
	getSampleIdsOnly bool) (map[string]interface{}, error) {

	// begin building the request body.
	mustMap := []map[string]interface{}{{
		"query_string": map[string]interface{}{
			"query": "chrom:" + chromosome,
		}},
	}
	shouldMap := []map[string]interface{}{}
	var minimumShouldMatch int

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

	if alternative != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": "alt:" + alternative,
			}})
	}

	if len(alleles) > 0 {
		switch len(alleles) {
		case 1:
			// any allele can be present on either side of the pair
			shouldMap = append(shouldMap, map[string]interface{}{
				"query_string": map[string]interface{}{
					"query": "sample.variation.alleles.left.keyword:" + alleles[0],
				}})
			shouldMap = append(shouldMap, map[string]interface{}{
				"query_string": map[string]interface{}{
					"query": "sample.variation.alleles.right.keyword:" + alleles[0],
				}})
			minimumShouldMatch = 1
		case 2:
			// treat as a left/right pair
			shouldMap = append(shouldMap, map[string]interface{}{
				"query_string": map[string]interface{}{
					"query": "sample.variation.alleles.left.keyword:" + alleles[0],
				}})
			shouldMap = append(shouldMap, map[string]interface{}{
				"query_string": map[string]interface{}{
					"query": "sample.variation.alleles.right.keyword:" + alleles[1],
				}})
			minimumShouldMatch = 2
		}
	}

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

	if tableId != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": "tableId:" + tableId,
			}})
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
		es.Search.WithIndex(variantsIndex),
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

func CountDocumentsContainerVariantOrSampleIdInPositionRange(cfg *models.Config, es *elasticsearch.Client,
	chromosome string, lowerBound int, upperBound int,
	variantId string, sampleId string,
	reference string, alternative string,
	genotype c.GenotypeQuery, assemblyId c.AssemblyId, tableId string) (map[string]interface{}, error) {

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

	if alternative != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": "alt:" + alternative,
			}})
	}

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

	if tableId != "" {
		mustMap = append(mustMap, map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": "tableId:" + tableId,
			}})
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
						"must": mustMap,
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
		es.Count.WithIndex(variantsIndex),
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

func GetVariantsBucketsByKeywordAndTableId(cfg *models.Config, es *elasticsearch.Client, keyword string, tableId string) (map[string]interface{}, error) {
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

	if tableId != "" {
		aggMap["query"] = map[string]interface{}{
			"match": map[string]interface{}{
				"tableId": tableId,
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

	if cfg.Debug {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	// Perform the search request.
	res, searchErr := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(variantsIndex),
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

func DeleteVariantsByTableId(es *es7.Client, cfg *models.Config, tableId string) (map[string]interface{}, error) {

	// cfg := c.(*contexts.GohanContext).Config
	// es := c.(*contexts.GohanContext).Es7Client

	if cfg.Debug {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"tableId": tableId,
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
		[]string{variantsIndex},
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
		fmt.Printf("Error unmarshalling gene search response: %s\n", umErr)
		return nil, umErr
	}

	return result, nil
}

// -- internal use only --

func addZygosityToMustMap(genotype c.GenotypeQuery, mustMap []map[string]interface{}) []map[string]interface{} {
	zygosityMatchMap := make(map[string]interface{})

	switch genotype {
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

	// Not all genotype-queries are compatible yet, i.e. Haploid types `REFERENCE` and `ALTERNATE`
	// - verify zygosity-map is not empty before adding to the must-map
	if len(zygosityMatchMap) > 0 {
		mustMap = append(mustMap, map[string]interface{}{
			"match": zygosityMatchMap,
		})
	}

	return mustMap
}
