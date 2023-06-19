package api

import (
	"encoding/json"
	"fmt"
	common "gohan/api/tests/common"

	"gohan/api/models"
	c "gohan/api/models/constants"
	a "gohan/api/models/constants/assembly-id"
	ingest "gohan/api/models/ingest"

	"gohan/api/models/constants/chromosome"
	"gohan/api/models/dtos"
	"gohan/api/models/indexes"

	"gohan/api/utils"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/ahmetb/go-linq"
)

const (
	GenesOverviewPath              string = "%s/genes/overview"
	GenesIngestionRunPath          string = "%s/genes/ingestion/run"
	GenesIngestionRequestsPath     string = "%s/genes/ingestion/requests"
	GenesSearchPathWithQueryString string = "%s/genes/search%s"
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

	// verify demo vcf was properly ingested
	t.Run("Test Genes Overview", func(t *testing.T) {
		// check variants overview
		cfg := common.InitConfig()

		overviewJson := getGenesOverview(t, cfg)
		assert.NotNil(t, overviewJson)
	})

	t.Run("Test Get Genes By AssemblyId And Chromosome", func(t *testing.T) {
		// retrieve all possible combinations of responses
		allDtoResponses := getAllDtosOfVariousCombinationsOfGenesAndAssemblyIDs(t)

		// assert the dto response slice is plentiful
		assert.NotNil(t, allDtoResponses)

		From(allDtoResponses).ForEachT(func(dto dtos.GenesResponseDTO) {
			// ensure there are results in the response
			assert.NotNil(t, dto.Results)

			// check the resulting data
			From(dto.Results).ForEachT(func(gene indexes.Gene) {
				// ensure the gene is legit
				assert.NotNil(t, gene.Name)
				assert.NotNil(t, gene.AssemblyId)
				assert.True(t, chromosome.IsValidHumanChromosome(gene.Chrom))
				assert.Greater(t, gene.End, gene.Start)
			})
		})
	})

}

func getGenesOverview(_t *testing.T, _cfg *models.Config) map[string]interface{} {
	request, _ := http.NewRequest("GET", fmt.Sprintf(GenesOverviewPath, _cfg.Api.Url), nil)

	client := &http.Client{}
	response, responseErr := client.Do(request)
	assert.Nil(_t, responseErr)

	defer response.Body.Close()

	// this test (at the time of writing) will only work if authorization is disabled
	shouldBe := 200
	assert.Equal(_t, shouldBe, response.StatusCode, fmt.Sprintf("Error -- Api GET / Status: %s ; Should be %d", response.Status, shouldBe))

	//	-- interpret array of ingestion requests from response
	overviewRespBody, overviewRespBodyErr := ioutil.ReadAll(response.Body)
	assert.Nil(_t, overviewRespBodyErr)

	//	--- transform body bytes to string
	overviewRespBodyString := string(overviewRespBody)

	//	-- check for json error
	var overviewRespJson map[string]interface{}
	overviewJsonUnmarshallingError := json.Unmarshal([]byte(overviewRespBodyString), &overviewRespJson)
	assert.Nil(_t, overviewJsonUnmarshallingError)

	// -- insure it's an empty array
	assemblyIDsKey, assidkOk := overviewRespJson["assemblyIDs"]
	assert.True(_t, assidkOk)
	assert.NotNil(_t, assemblyIDsKey)

	return overviewRespJson
}

func getAllDtosOfVariousCombinationsOfGenesAndAssemblyIDs(_t *testing.T) []dtos.GenesResponseDTO {
	cfg := common.InitConfig()

	// retrieve the overview
	overviewJson := getGenesOverview(_t, cfg)

	// ensure the response is valid
	// TODO: error check instead of nil check
	assert.NotNil(_t, overviewJson)

	// initialize a common slice in which to
	// accumulate al responses asynchronously
	allDtoResponses := []dtos.GenesResponseDTO{}
	allDtoResponsesMux := sync.RWMutex{}

	var combWg sync.WaitGroup
	for _, assemblyIdOverviewBucket := range overviewJson {

		// range over all assembly IDs
		for assemblyIdString, genesPerChromosomeBucket := range assemblyIdOverviewBucket.(map[string]interface{}) {

			fmt.Println(assemblyIdString)
			fmt.Println(genesPerChromosomeBucket)

			castedBucket := genesPerChromosomeBucket.(map[string]interface{})["numberOfGenesPerChromosome"].(map[string]interface{})

			for chromosomeString, _ := range castedBucket { // _ = number of genes (unused)

				combWg.Add(1)
				go func(_wg *sync.WaitGroup, _assemblyIdString string, _chromosomeString string) {
					defer _wg.Done()

					assemblyId := a.CastToAssemblyId(_assemblyIdString)

					// make the call
					dto := buildQueryAndMakeGetGenesCall(_chromosomeString, "", assemblyId, _t, cfg)

					// ensure there is data returned
					// (we'd be making a bad query, otherwise)
					assert.True(_t, len(dto.Results) > 0)

					// accumulate all response objects
					// to a common slice in an
					// asynchronous-safe manner
					allDtoResponsesMux.Lock()
					allDtoResponses = append(allDtoResponses, dto)
					allDtoResponsesMux.Unlock()
				}(&combWg, assemblyIdString, chromosomeString)
			}

		}

	}
	combWg.Wait()

	return allDtoResponses
}

func buildQueryAndMakeGetGenesCall(chromosome string, term string, assemblyId c.AssemblyId, _t *testing.T, _cfg *models.Config) dtos.GenesResponseDTO {

	queryString := fmt.Sprintf("?chromosome=%s&assemblyId=%s", chromosome, assemblyId)

	url := fmt.Sprintf(GenesSearchPathWithQueryString, _cfg.Api.Url, queryString)

	return getGetGenesCall(url, _t)
}

func getGetGenesCall(url string, _t *testing.T) dtos.GenesResponseDTO {
	fmt.Printf("Calling %s\n", url)
	request, _ := http.NewRequest("GET", url, nil)

	client := &http.Client{}
	response, responseErr := client.Do(request)
	assert.Nil(_t, responseErr)

	defer response.Body.Close()

	// this test (at the time of writing) will only work if authorization is disabled
	shouldBe := 200
	assert.Equal(_t, shouldBe, response.StatusCode, fmt.Sprintf("Error -- Api GET %s Status: %s ; Should be %d", url, response.Status, shouldBe))

	//	-- interpret array of ingestion requests from response
	respBody, respBodyErr := ioutil.ReadAll(response.Body)
	assert.Nil(_t, respBodyErr)

	//	--- transform body bytes to string
	respBodyString := string(respBody)

	//	-- convert to json and check for error
	var respDto dtos.GenesResponseDTO
	jsonUnmarshallingError := json.Unmarshal([]byte(respBodyString), &respDto)
	assert.Nil(_t, jsonUnmarshallingError)

	return respDto
}
