package api

import (
	"fmt"
	ingest "gohan/api/models/ingest"
	common "gohan/api/tests/common"
	"gohan/api/utils"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	VariantsOverviewPath                      string = "%s/variants/overview"
	VariantsGetBySampleIdsPathWithQueryString string = "%s/variants/get/by/sampleId%s"
	IngestionRequestsPath                     string = "%s/variants/ingestion/requests"
)

func TestIngest(t *testing.T) {
	cfg := common.InitConfig()
	tableId := uuid.NewString()

	assert.True(t, t.Run("Ingest Demo VCF", func(t *testing.T) {
		// create demo vcf string
		sampleId := "abc1234"

		// - save string to vcf directory
		localDataRootPath := common.GetRootGohanPath() + "/data"
		localVcfPath := localDataRootPath + "/vcfs"

		newFilePath := fmt.Sprintf("%s/%s.vcf", localVcfPath, sampleId)

		// - create file if not exists
		file, err := common.CreateAndGetNewFile(newFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			file.Close()
		}()

		// - write some vcf string to file.
		_, err = file.WriteString(common.DemoVcf1)
		if common.IsError(err) {
			return
		}
		defer func() {
			os.Remove(newFilePath)
		}()

		// compress the vcf file with bgzip
		out, err := exec.Command("bgzip", newFilePath).Output()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(string(out))

		newGzFile := newFilePath + ".gz"
		defer func() {
			os.Remove(newGzFile)
		}()

		// - ingest
		assemblyId := "GRCh38"
		containerizedVcfFilePath := "/data/" + filepath.Base(newGzFile)

		queryString := fmt.Sprintf("assemblyId=%s&fileNames=%s&tableId=%s", assemblyId, containerizedVcfFilePath, tableId)
		ingestUrl := fmt.Sprintf("%s/variants/ingestion/run?%s", cfg.Api.Url, queryString)

		initialIngestionDtos := utils.GetRequestReturnStuff[[]ingest.IngestResponseDTO](ingestUrl)
		assert.True(t, len(initialIngestionDtos) > 0)

		// check ingestion request
		// TODO: avoid potential infinite loop
		for {
			fmt.Println("Checking state of the ingestion..")

			// make the call
			ingReqsUrl := fmt.Sprintf("%s/variants/ingestion/requests", cfg.Api.Url)
			ingReqDtos := utils.GetRequestReturnStuff[[]ingest.IngestResponseDTO](ingReqsUrl)
			assert.True(t, len(initialIngestionDtos) > 0)

			foundDone := false
			for _, dto := range ingReqDtos {
				if dto.Filename == filepath.Base(containerizedVcfFilePath) && dto.State == "Done" {
					foundDone = true
					break
				}
				if dto.Filename == filepath.Base(containerizedVcfFilePath) && dto.State == "Error" {
					log.Fatal(dto.Message)
				}
			}
			if foundDone {
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
			statsReqUrl := fmt.Sprintf("%s/variants/ingestion/stats", cfg.Api.Url)
			stats := utils.GetRequestReturnStuff[ingest.IngestStatsDto](statsReqUrl)
			assert.NotNil(t, stats)

			fmt.Println(stats.NumAdded)
			fmt.Println(stats.NumFlushed)
			if stats.NumAdded == stats.NumFlushed {
				fmt.Println("Done, moving on..")
				break
			}
			if stats.NumFailed > 0 {
				log.Fatal("More than one variant failed to flush")
			}

			// pause
			time.Sleep(3 * time.Second)
		}
	}))

	// verify demo vcf was properly ingested
	// by pinging it with specific queries
	assert.True(t, t.Run("Check Demo VCF Ingestion", func(t *testing.T) {
		overviewJson := common.GetVariantsOverview(t, cfg)
		assert.NotNil(t, overviewJson)

		dtos := common.BuildQueryAndMakeGetVariantsCall("1", "*", true, "asc", "", "GRCh38", "", "", "", false, t, cfg)
		assert.True(t, len(dtos.Results[0].Calls) > 0)
	}))
}
