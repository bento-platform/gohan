package variantsService

import (
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

func GetVariantsOverview(es *elasticsearch.Client, cfg *models.Config) map[string]interface{} {
	resultsMap := map[string]interface{}{}
	resultsMux := sync.RWMutex{}

	var wg sync.WaitGroup
	callGetBucketsByKeyword := func(key string, keyword string, _wg *sync.WaitGroup) {
		defer _wg.Done()

		results, bucketsError := esRepo.GetVariantsBucketsByKeywordAndTableId(cfg, es, keyword, "")
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
		if aggs, aggsOk := results["aggregations"]; aggsOk {
			aggsMapped := aggs.(map[string]interface{})

			if items, itemsOk := aggsMapped["items"]; itemsOk {
				itemsMapped := items.(map[string]interface{})

				if buckets, bucketsOk := itemsMapped["buckets"]; bucketsOk {
					bucketsMapped = buckets.([]interface{})
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

	// get distribution of table IDs
	wg.Add(1)
	go callGetBucketsByKeyword("tableIDs", "tableId.keyword", &wg)

	wg.Wait()

	return resultsMap
}
