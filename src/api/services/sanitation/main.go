package sanitation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	es7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/go-co-op/gocron"
	"github.com/mitchellh/mapstructure"

	"api/models"
	"api/models/indexes"
	esRepo "api/repositories/elasticsearch"

	variantsService "api/services/variants"
)

type (
	SanitationService struct {
		Initialized bool
		Es7Client   *es7.Client
		Config      *models.Config
	}
)

func NewSanitationService(es *es7.Client, cfg *models.Config) *SanitationService {
	ss := &SanitationService{
		Initialized: false,
		Es7Client:   es,
		Config:      cfg,
	}

	ss.Init()

	return ss
}

func (ss *SanitationService) Init() {
	// initialization if necessary
	if !ss.Initialized {
		// - spin up a go routine that will periodically
		//   run through a series of steps to ensure
		//   the system is "sanitary" ; i.e. in an elasticsearch
		//   context, that would mean performing something like
		//   - removing duplicate documents
		//   - cleaning documents that have broken pseudo-foreign keys
		//     - variants -> tables
		//   etc...
		go func() {
			// setup cron job
			s := gocron.NewScheduler(time.UTC)

			// clean variant documents with non-existing tables
			s.Every(1).Days().At("04:00:00").Do(func() { // 12am EST
				fmt.Printf("[%s] - Running variant documents cleanup..\n", time.Now())

				// - get all available tables
				tables, tablesError := esRepo.GetTables(ss.Config, ss.Es7Client, context.Background(), "", "variant")
				if tablesError != nil {
					fmt.Printf("[%s] - Error getting tables : %v..\n", time.Now(), tablesError)
					return
				}

				// gather data from "hits"
				docsHits := tables["hits"].(map[string]interface{})["hits"]
				allDocHits := []map[string]interface{}{}
				mapstructure.Decode(docsHits, &allDocHits)

				// grab _source for each hit
				tableIds := make([]string, 0)
				for _, r := range allDocHits {
					source := r["_source"]
					byteSlice, _ := json.Marshal(source)

					// cast map[string]interface{} to table
					var resultingTable indexes.Table
					if err := json.Unmarshal(byteSlice, &resultingTable); err != nil {
						fmt.Println("failed to unmarshal:", err)
						continue
					}

					// accumulate structs
					tableIds = append(tableIds, resultingTable.Id)
				}
				fmt.Printf("[%s] - Table IDs found : %v..\n", time.Now(), tableIds)

				// - obtain distribution of table IDs accross all variants
				// TODO: refactor not use variants-mvc package to access this (anti-pattern)
				variantsOverview := variantsService.GetVariantsOverview(ss.Es7Client, ss.Config)
				if variantsOverview == nil {
					return
				}
				if variantsOverview["tableIDs"] == nil {
					return
				}

				variantTableIdsCountsMap := variantsOverview["tableIDs"].(map[string]interface{})

				variantTableIds := make([]string, 0)
				for _variantTableId, _ := range variantTableIdsCountsMap {
					// ignore variant count set to _

					// accumulate IDs found in list
					variantTableIds = append(variantTableIds, _variantTableId)
				}
				fmt.Printf("[%s] - Tables IDs found across all variants : %v..\n", time.Now(), variantTableIds)

				// obtain set-difference between variant-table IDs table IDs
				setDiff := setDifference(tableIds, variantTableIds)
				fmt.Printf("[%s] - Variant Table ID Difference: %v..\n", time.Now(), setDiff)

				// delete variants with table IDs found in this set difference
				for _, differentId := range setDiff {
					// TODO: refactor
					// fire and forget
					go func(_differentId string) {
						_, _ = esRepo.DeleteVariantsByTableId(ss.Es7Client, ss.Config, _differentId)
					}(differentId)
				}
			})

			// starts the scheduler in blocking mode, which blocks
			// the current execution path
			s.StartBlocking()
		}()

		ss.Initialized = true
		fmt.Println("Sanitation Service Initialized ..")
	}
}

func setDifference(a, b []string) (c []string) {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			c = append(c, item)
		}
	}
	return
}
