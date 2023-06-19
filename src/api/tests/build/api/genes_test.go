package api

import (
	"fmt"
	ingest "gohan/api/models/ingest"
	common "gohan/api/tests/common"
	"gohan/api/utils"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	GenesOverviewPath          string = "%s/genes/overview"
	GenesIngestionRunPath      string = "%s/genes/ingestion/run"
	GenesIngestionRequestsPath string = "%s/genes/ingestion/requests"
)

func TestGenesIngestion(t *testing.T) {
	cfg := common.InitConfig()

	t.Run("Ingest And Check Genes", func(t *testing.T) {
		// - ingest
		ingestUrl := fmt.Sprintf(GenesIngestionRunPath, cfg.Api.Url)

		initialIngestionDtos := utils.GetRequestReturnStuff[ingest.GeneIngestRequest](ingestUrl)
		assert.True(t, len(initialIngestionDtos.Message) > 0)

		// check ingestion request
		// TODO: avoid potential infinite loop
		for {
			fmt.Println("Checking state of the ingestion..")

			// make the call
			ingReqsUrl := fmt.Sprintf(GenesIngestionRequestsPath, cfg.Api.Url)
			ingReqDtos := utils.GetRequestReturnStuff[[]ingest.GeneIngestRequest](ingReqsUrl)
			assert.True(t, len(ingReqDtos) > 0)

			numFilesDone := 0
			numFilesRunning := len(ingReqDtos)
			for _, dto := range ingReqDtos {
				if dto.State == "Done" {
					numFilesDone += 1
				}
				if dto.State == "Error" {
					log.Fatal(dto.Message)
				}
			}
			if numFilesDone == numFilesRunning {
				fmt.Println("Done, moving on..")
				break
			} else {
				// pause
				time.Sleep(3 * time.Second)
			}
		}

		// check ingestion stats
		// TODO: avoid potential infinite loop
		for {
			fmt.Println("Checking ingestion stats..")
			// pause
			time.Sleep(3 * time.Second)

			// make the call
			statsReqUrl := fmt.Sprintf("%s/genes/ingestion/stats", cfg.Api.Url)
			stats := utils.GetRequestReturnStuff[ingest.IngestStatsDto](statsReqUrl)
			assert.NotNil(t, stats)

			fmt.Println(stats.NumAdded)
			fmt.Println(stats.NumFlushed)
			if stats.NumAdded == stats.NumFlushed {
				fmt.Println("Done, moving on..")
				break
			}
			if stats.NumFailed > 0 {
				log.Fatal("More than one gene failed to flush")
			}

			// pause
			time.Sleep(3 * time.Second)
		}
	})
}
