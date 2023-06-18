package api

import (
	"fmt"
	"gohan/api/models/dtos"
	"gohan/api/models/indexes"
	ingest "gohan/api/models/ingest"
	common "gohan/api/tests/common"
	"gohan/api/utils"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	gq "gohan/api/models/constants/genotype-query"
	s "gohan/api/models/constants/sort"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	. "github.com/ahmetb/go-linq"
)

const (
	VariantsOverviewPath                      string = "%s/variants/overview"
	VariantsGetBySampleIdsPathWithQueryString string = "%s/variants/get/by/sampleId%s"
	IngestionRequestsPath                     string = "%s/variants/ingestion/requests"
)

func TestDemoVcfIngestion(t *testing.T) {
	cfg := common.InitConfig()
	tableId := uuid.NewString()

	t.Run("Ingest Demo VCF", func(t *testing.T) {
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
	})

	// verify demo vcf was properly ingested
	t.Run("Test Variants Overview", func(t *testing.T) {
		// check variants overview
		overviewJson := common.GetVariantsOverview(t, cfg)
		assert.NotNil(t, overviewJson)
	})

	t.Run("Test Simple Chromosome Queries", func(t *testing.T) {
		// simple chromosome-1 query
		chromQueryResponse := common.BuildQueryAndMakeGetVariantsCall("1", "*", true, "asc", "", "GRCh38", "", "", "", false, t, cfg)
		assert.True(t, len(chromQueryResponse.Results[0].Calls) > 0)
	})

	t.Run("Test Simple Allele Queries", func(t *testing.T) {
		// TODO: not hardcoded tests
		// simple allele queries
		common.GetAndVerifyVariantsResults(cfg, t, "CAG")
		common.GetAndVerifyVariantsResults(cfg, t, "CAAAA")
		common.GetAndVerifyVariantsResults(cfg, t, "T")
		common.GetAndVerifyVariantsResults(cfg, t, "C")

		// random number between 1 and 5
		// allelleLen := rand.Intn(5) + 1

		// random nucleotide string of length 'allelleLen'
		// qAllele := utils.GenerateRandomFixedLengthString(utils.AcceptedNucleotideCharacters, allelleLen)
	})

	t.Run("Test Variant Info Present", func(t *testing.T) {
		allDtoResponses := common.GetAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(t, true, s.Undefined, gq.UNCALLED, "", "")

		// assert that all of the responses include valid sets of info
		// - * accumulate all infos into a single list using the set of
		//   SelectManyT's and the SelectT
		// - ** iterate over each info in the ForEachT
		var accumulatedInfos []*indexes.Info

		From(allDtoResponses).SelectManyT(func(resp dtos.VariantGetReponse) Query { // *
			return From(resp.Results)
		}).SelectManyT(func(data dtos.VariantGetResult) Query {
			return From(data.Calls)
		}).SelectManyT(func(variant dtos.VariantCall) Query {
			return From(variant.Info)
		}).SelectT(func(info indexes.Info) indexes.Info {
			return info
		}).ForEachT(func(info indexes.Info) { // **
			accumulatedInfos = append(accumulatedInfos, &info)
		})

		if len(accumulatedInfos) == 0 {
			t.Skip("No infos returned! Skipping --")
		}

		for infoIndex, info := range accumulatedInfos {
			// ensure the info is not nil
			// - s.Id can be == ""
			// - so can s.Value
			assert.NotNil(t, info)
			if info.Id == "" {
				fmt.Printf("Note: Found empty info id at index %d with value %s \n", infoIndex, info.Value)
			}
		}
	})

	t.Run("Test No Variant Info Present", func(t *testing.T) {

		allDtoResponses := common.GetAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(t, false, s.Undefined, gq.UNCALLED, "", "")

		// assert that all responses from all combinations have no results
		for _, dtoResponse := range allDtoResponses {
			if len(dtoResponse.Results) > 0 {
				firstDataPointCalls := dtoResponse.Results[0].Calls
				if len(firstDataPointCalls) > 0 {
					assert.Nil(t, firstDataPointCalls[0].Info)
				}
			}
		}
	})

	t.Run("Test Get Variants in Ascending Order", func(t *testing.T) {
		// retrieve responses in ascending order
		allDtoResponses := common.GetAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(t, false, s.Ascending, gq.UNCALLED, "", "")

		// assert the dto response slice is plentiful
		assert.NotNil(t, allDtoResponses)

		From(allDtoResponses).ForEachT(func(dto dtos.VariantGetReponse) {
			// ensure there is data
			assert.NotNil(t, dto.Results)

			// check the data
			From(dto.Results).ForEachT(func(d dtos.VariantGetResult) {
				// ensure the variants slice is plentiful
				assert.NotNil(t, d.Calls)

				latestSmallest := 0
				From(d.Calls).ForEachT(func(dd dtos.VariantCall) {
					// verify order
					if latestSmallest != 0 {
						assert.True(t, latestSmallest <= dd.Pos)
					}

					latestSmallest = dd.Pos
				})
			})
		})
	})

	t.Run("Test Get Variants in Descending Order", func(t *testing.T) {
		// retrieve responses in descending order
		allDtoResponses := common.GetAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(t, false, s.Descending, gq.UNCALLED, "", "")

		// assert the dto response slice is plentiful
		assert.NotNil(t, allDtoResponses)

		From(allDtoResponses).ForEachT(func(dto dtos.VariantGetReponse) {
			// ensure there is data
			assert.NotNil(t, dto.Results)

			// check the data
			From(dto.Results).ForEachT(func(d dtos.VariantGetResult) {
				// ensure the variants slice is plentiful
				assert.NotNil(t, d.Calls)

				latestGreatest := 0
				From(d.Calls).ForEachT(func(dd dtos.VariantCall) {
					if latestGreatest != 0 {
						assert.True(t, latestGreatest >= dd.Pos)
					}

					latestGreatest = dd.Pos
				})
			})
		})
	})
}
