package mvc

import (
	"api/contexts"
	"api/models"
	assemblyId "api/models/constants/assembly-id"
	esRepo "api/repositories/elasticsearch"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/labstack/echo"
	"github.com/mitchellh/mapstructure"
)

func GenesGetByNomenclatureWildcard(c echo.Context) error {
	cfg := c.(*contexts.GohanContext).Config
	es := c.(*contexts.GohanContext).Es7Client

	// Name search term
	term := c.QueryParam("term")

	// Assembly ID
	// perform wildcard search if empty/random parameter is passed
	// - set to Unknown to trigger it
	assId := assemblyId.Unknown
	if assemblyId.CastToAssemblyId(c.QueryParam("assemblyId")) != assemblyId.Unknown {
		// retrieve passed parameter if is valid
		assId = assemblyId.CastToAssemblyId(c.QueryParam("assemblyId"))
	}

	// Size
	var (
		size        int = 25
		sizeCastErr error
	)
	if len(c.QueryParam("size")) > 0 {
		sizeQP := c.QueryParam("size")
		size, sizeCastErr = strconv.Atoi(sizeQP)
		if sizeCastErr != nil {
			size = 25
		}
	}

	fmt.Printf("Executing wildcard genes search for term %s, assemblyId %s (max size: %d)\n", term, assId, size)

	// Execute
	docs := esRepo.GetGeneDocumentsByTermWildcard(cfg, es, term, assId, size)

	docsHits := docs["hits"].(map[string]interface{})["hits"]
	allDocHits := []map[string]interface{}{}
	mapstructure.Decode(docsHits, &allDocHits)

	// grab _source for each hit
	var allSources []models.Gene

	for _, r := range allDocHits {
		source := r["_source"].(map[string]interface{})

		// cast map[string]interface{} to struct
		var resultingVariant models.Gene
		mapstructure.Decode(source, &resultingVariant)

		// accumulate structs
		allSources = append(allSources, resultingVariant)
	}

	fmt.Printf("Found %d docs!\n", len(allSources))

	geneResponseDTO := models.GenesResponseDTO{
		Term:    term,
		Count:   len(allSources),
		Results: allSources,
		Status:  200,
		Message: "Success",
	}

	return c.JSON(http.StatusOK, geneResponseDTO)
}

func GetGenesOverview(c echo.Context) error {

	resultsMap := map[string]interface{}{}
	resultsMux := sync.RWMutex{}

	var wg sync.WaitGroup
	es := c.(*contexts.GohanContext).Es7Client
	cfg := c.(*contexts.GohanContext).Config

	callGetBucketsByKeyword := func(key string, keyword string, _wg *sync.WaitGroup) {
		defer _wg.Done()

		results := esRepo.GetGeneBucketsByKeyword(cfg, es, keyword)

		// retrieve aggregations.items.buckets
		bucketsMapped := []interface{}{}
		if aggs, ok := results["aggregations"]; ok {
			aggsMapped := aggs.(map[string]interface{})

			if items, ok := aggsMapped["items"]; ok {
				itemsMapped := items.(map[string]interface{})

				if buckets := itemsMapped["buckets"]; ok {
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
	go callGetBucketsByKeyword("chromosomes", "chrom", &wg)

	// get distribution of variant IDs
	wg.Add(1)
	go callGetBucketsByKeyword("assemblyIDs", "assemblyId.keyword", &wg)

	wg.Wait()

	return c.JSON(http.StatusOK, resultsMap)
}
