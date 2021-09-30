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

	// Chromosome search term
	chromosome := c.QueryParam("chromosome")

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
	docs := esRepo.GetGeneDocumentsByTermWildcard(cfg, es, chromosome, term, assId, size)

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

	es := c.(*contexts.GohanContext).Es7Client
	cfg := c.(*contexts.GohanContext).Config

	// retrieve aggregation of genes/chromosomes by assembly id
	results := esRepo.GetGeneBucketsByKeyword(cfg, es)

	// begin mapping results
	geneChromosomeGroupBucketsMapped := []map[string]interface{}{}

	// loop over top level aggregation and
	// accumulated nested aggregations
	if aggs, ok := results["aggregations"]; ok {
		aggsMapped := aggs.(map[string]interface{})

		if items, ok := aggsMapped["genes_assembly_id_group"]; ok {
			itemsMapped := items.(map[string]interface{})

			if buckets := itemsMapped["buckets"]; ok {
				arrayMappedBuckets := buckets.([]interface{})

				for _, mappedBucket := range arrayMappedBuckets {
					geneChromosomeGroupBucketsMapped = append(geneChromosomeGroupBucketsMapped, mappedBucket.(map[string]interface{}))
				}
			}
		}
	}

	individualAssemblyIdKeyMap := map[string]interface{}{}

	// iterated over each assemblyId bucket
	for _, chromGroupBucketMap := range geneChromosomeGroupBucketsMapped {

		assemblyIdKey := fmt.Sprint(chromGroupBucketMap["key"])

		numGenesPerChromMap := map[string]interface{}{}
		bucketsMapped := map[string]interface{}{}

		if chromGroupItem, ok := chromGroupBucketMap["genes_chromosome_group"]; ok {
			chromGroupItemMapped := chromGroupItem.(map[string]interface{})

			for _, chromBucket := range chromGroupItemMapped["buckets"].([]interface{}) {
				doc_key := fmt.Sprint(chromBucket.(map[string]interface{})["key"]) // ensure strings and numbers are expressed as strings
				doc_count := chromBucket.(map[string]interface{})["doc_count"]

				// add to list of buckets by chromosome
				bucketsMapped[doc_key] = doc_count
			}
		}

		numGenesPerChromMap["numberOfGenesPerChromosome"] = bucketsMapped
		individualAssemblyIdKeyMap[assemblyIdKey] = numGenesPerChromMap
	}

	resultsMux.Lock()
	resultsMap["assemblyIDs"] = individualAssemblyIdKeyMap
	resultsMux.Unlock()

	return c.JSON(http.StatusOK, resultsMap)
}
