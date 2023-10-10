package variantsService

import (
	"errors"
	"fmt"
	"gohan/api/models"
	esRepo "gohan/api/repositories/elasticsearch"
	"sync"

	"github.com/elastic/go-elasticsearch/v7"
)

type (
	VariantService struct {
		Config *models.Config
	}
)

func NewVariantService(cfg *models.Config) *VariantService {
	vs := &VariantService{
		Config: cfg,
	}

	return vs
}

func GetVariantsOverview(es *elasticsearch.Client, cfg *models.Config) (map[string]interface{}, error) {
	resultsMap := map[string]interface{}{}
	resultsMux := sync.RWMutex{}

	var wg sync.WaitGroup
	callGetBucketsByKeyword := func(key string, keyword string, _wg *sync.WaitGroup) {
		defer _wg.Done()

		results, bucketsError := esRepo.GetVariantsBucketsByKeyword(cfg, es, keyword)
		if bucketsError != nil {
			resultsMux.Lock()
			defer resultsMux.Unlock()

			resultsMap[key] = map[string]interface{}{
				"error": "Something went wrong. Please contact the administrator!",
			}
			return
		}

		// retrieve aggregations.items.buckets
		bucketsMapped := []interface{}{}
		if aggs, aggsOk := results["aggregations"].(map[string]interface{}); aggsOk {
			if latest, latestOk := aggs["latest_created"].(map[string]interface{}); latestOk {
				if valueAsString, valOk := latest["value_as_string"].(string); valOk {
					resultsMux.Lock()
					resultsMap["last_created_time"] = valueAsString
					resultsMux.Unlock()
				}
			}
			if items, itemsOk := aggs["items"].(map[string]interface{}); itemsOk {
				if buckets, bucketsOk := items["buckets"].([]interface{}); bucketsOk {
					bucketsMapped = buckets
				}
			}
		}

		individualKeyMap := map[string]interface{}{}
		// push results bucket to slice
		for _, bucket := range bucketsMapped {
			doc_key := fmt.Sprint(bucket.(map[string]interface{})["key"]) // ensure strings and numbers are expressed as strings
			doc_count := bucket.(map[string]interface{})["doc_count"]

			individualKeyMap[doc_key] = doc_count
		}

		resultsMux.Lock()
		resultsMap[key] = individualKeyMap
		resultsMux.Unlock()
	}

	// First, make sure the ES cluster is running - otherwise this will hang for a long time
	_, err := es.Ping()
	if err != nil {
		return nil, errors.New("could not contact Elasticsearch - make sure it's running")
	}

	// Extract latest created time
	if latest, exists := resultsMap["last_ingested"].(map[string]interface{}); exists {
		latestCreatedTime := latest["value_as_string"].(string)
		resultsMap["last_created_time"] = latestCreatedTime
	}

	// get distribution of chromosomes
	wg.Add(1)
	go callGetBucketsByKeyword("chromosomes", "chrom.keyword", &wg)

	// get distribution of variant IDs
	wg.Add(1)
	go callGetBucketsByKeyword("variantIDs", "id.keyword", &wg)

	// get distribution of sample IDs
	wg.Add(1)
	go callGetBucketsByKeyword("sampleIDs", "sample.id.keyword", &wg)

	// get distribution of assembly IDs
	wg.Add(1)
	go callGetBucketsByKeyword("assemblyIDs", "assemblyId.keyword", &wg)

	// get distribution of datasets
	wg.Add(1)
	go callGetBucketsByKeyword("datasets", "dataset.keyword", &wg)

	wg.Wait()

	return resultsMap, nil
}
