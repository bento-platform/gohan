package api

import (
	"fmt"
	"gohan/api/models/dtos"
	"gohan/api/models/indexes"
	ingest "gohan/api/models/ingest"
	common "gohan/api/tests/common"
	"gohan/api/utils"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	c "gohan/api/models/constants"
	a "gohan/api/models/constants/assembly-id"
	gq "gohan/api/models/constants/genotype-query"
	s "gohan/api/models/constants/sort"
	z "gohan/api/models/constants/zygosity"
	ratt "gohan/api/tests/common/constants/referenceAlternativeTestType"

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
		// verify ingestion endpoint
		// -- ensure nothing is running
		initialIngestionState := utils.GetRequestReturnStuff[[]ingest.IngestResponseDTO](fmt.Sprintf(common.IngestionRequestsPath, cfg.Api.Url))
		assert.NotNil(t, len(initialIngestionState))

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

		// pause
		time.Sleep(1 * time.Second)

		// check ingestion request
		// TODO: avoid potential infinite loop
		counter := 0
		for {
			fmt.Printf("\rChecking state of the variants ingestion.. [%d]", counter)

			// make the call
			ingReqsUrl := fmt.Sprintf("%s/variants/ingestion/requests", cfg.Api.Url)
			ingReqDtos := utils.GetRequestReturnStuff[[]ingest.IngestResponseDTO](ingReqsUrl)
			assert.True(t, len(ingReqDtos) > 0)

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
				fmt.Println("\nDone, moving on..")
				break
			} else {
				// pause
				time.Sleep(3 * time.Second)
			}
			counter++
		}

		// check ingestion stats
		// TODO: avoid potential infinite loop
		counter = 0
		for {
			fmt.Printf("\rChecking ingestion stats.. [%d]", counter)
			// pause
			time.Sleep(3 * time.Second)

			// make the call
			statsReqUrl := fmt.Sprintf("%s/variants/ingestion/stats", cfg.Api.Url)
			stats := utils.GetRequestReturnStuff[ingest.IngestStatsDto](statsReqUrl)
			assert.NotNil(t, stats)

			fmt.Println(stats.NumAdded)
			fmt.Println(stats.NumFlushed)
			if stats.NumAdded == stats.NumFlushed {
				fmt.Println("\nDone, moving on..")
				break
			}
			if stats.NumFailed > 0 {
				log.Fatal("\nMore than one variant failed to flush")
			}

			// pause
			time.Sleep(3 * time.Second)
			counter++
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
	validateReferenceSample := func(__t *testing.T, call *dtos.VariantCall) {
		assert.True(__t, call.GenotypeType == z.ZygosityToString(z.Reference))
	}

	validateAlternateSample := func(__t *testing.T, call *dtos.VariantCall) {
		assert.True(__t, call.GenotypeType == z.ZygosityToString(z.Alternate))
	}

	validateHeterozygousSample := func(__t *testing.T, call *dtos.VariantCall) {
		assert.True(__t, call.GenotypeType == z.ZygosityToString(z.Heterozygous))
	}

	validateHomozygousReferenceSample := func(__t *testing.T, call *dtos.VariantCall) {
		assert.True(__t, call.GenotypeType == z.ZygosityToString(z.HomozygousReference))
	}

	validateHomozygousAlternateSample := func(__t *testing.T, call *dtos.VariantCall) {
		assert.True(__t, call.GenotypeType == z.ZygosityToString(z.HomozygousAlternate))
	}
	t.Run("Test Get Variants Samples", func(t *testing.T) {

		// Reference Samples
		runAndValidateGenotypeQueryResults(t, gq.REFERENCE, validateReferenceSample)

		// Alternate Samples
		runAndValidateGenotypeQueryResults(t, gq.ALTERNATE, validateAlternateSample)

		// HeterozygousSamples
		runAndValidateGenotypeQueryResults(t, gq.HETEROZYGOUS, validateHeterozygousSample)

		// HomozygousReferenceSamples
		runAndValidateGenotypeQueryResults(t, gq.HOMOZYGOUS_REFERENCE, validateHomozygousReferenceSample)

		// Homozygous Alternate Samples
		runAndValidateGenotypeQueryResults(t, gq.HOMOZYGOUS_ALTERNATE, validateHomozygousAlternateSample)
	})
	t.Run("Test Get Variants Samples with Specific Alleles", func(t *testing.T) {
		// Homozygous Alternate Variants With Various References
		specificValidation := func(__t *testing.T, call *dtos.VariantCall, referenceAllelePattern string, alternativeAllelePattern string) {
			// ensure test is formatted correctly
			assert.True(__t, alternativeAllelePattern == "")

			// validate variant
			assert.Contains(__t, call.Ref, referenceAllelePattern)

			validateHomozygousAlternateSample(__t, call)
		}
		common.ExecuteReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HOMOZYGOUS_ALTERNATE, ratt.Reference, specificValidation)

		// Homozygous Reference Variants With Various References
		specificValidation = func(__t *testing.T, call *dtos.VariantCall, referenceAllelePattern string, alternativeAllelePattern string) {
			// ensure test is formatted correctly
			assert.True(__t, alternativeAllelePattern == "")

			// validate variant
			assert.Contains(__t, call.Ref, referenceAllelePattern)

			validateHomozygousReferenceSample(__t, call)
		}
		common.ExecuteReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HOMOZYGOUS_REFERENCE, ratt.Reference, specificValidation)

		//Heterozygous Variants With Various References
		specificValidation = func(__t *testing.T, call *dtos.VariantCall, referenceAllelePattern string, alternativeAllelePattern string) {
			// ensure test is formatted correctly
			assert.True(__t, alternativeAllelePattern == "")

			// validate variant
			assert.Contains(__t, call.Ref, referenceAllelePattern)

			validateHeterozygousSample(__t, call)
		}
		common.ExecuteReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HETEROZYGOUS, ratt.Reference, specificValidation)

		// Homozygous Alternate Variants With Various Alternatives
		specificValidation = func(__t *testing.T, call *dtos.VariantCall, referenceAllelePattern string, alternativeAllelePattern string) {
			// ensure test is formatted correctly
			assert.True(__t, referenceAllelePattern == "")

			// validate variant
			assert.Contains(__t, call.Alt, alternativeAllelePattern)

			validateHomozygousAlternateSample(__t, call)
		}
		common.ExecuteReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HOMOZYGOUS_ALTERNATE, ratt.Alternative, specificValidation)

		// Homozygous Reference Variants With Various Alternatives
		specificValidation = func(__t *testing.T, call *dtos.VariantCall, referenceAllelePattern string, alternativeAllelePattern string) {
			// ensure test is formatted correctly
			assert.True(__t, referenceAllelePattern == "")

			// validate variant
			assert.Contains(__t, call.Alt, alternativeAllelePattern)

			validateHomozygousReferenceSample(__t, call)
		}
		common.ExecuteReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HOMOZYGOUS_REFERENCE, ratt.Alternative, specificValidation)

		// Heterozygous Variants With Various Alternatives
		specificValidation = func(__t *testing.T, call *dtos.VariantCall, referenceAllelePattern string, alternativeAllelePattern string) {
			// ensure test is formatted correctly
			assert.True(__t, referenceAllelePattern == "")

			// validate variant
			assert.Contains(__t, call.Alt, alternativeAllelePattern)

			validateHeterozygousSample(__t, call)
		}
		common.ExecuteReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HETEROZYGOUS, ratt.Alternative, specificValidation)
	})
}

func runAndValidateGenotypeQueryResults(_t *testing.T, genotypeQuery c.GenotypeQuery, specificValidation func(__t *testing.T, call *dtos.VariantCall)) {

	allDtoResponses := getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(_t, true, s.Undefined, genotypeQuery, "", "")

	// assert that all of the responses include heterozygous sample sets
	// - * accumulate all samples into a single list using the set of SelectManyT's and the SelectT
	// - ** iterate over each sample in the ForEachT
	// var accumulatedSamples []*indexes.Sample
	var accumulatedCalls []*dtos.VariantCall

	From(allDtoResponses).SelectManyT(func(resp dtos.VariantGetReponse) Query { // *
		return From(resp.Results)
	}).SelectManyT(func(data dtos.VariantGetResult) Query {
		return From(data.Calls)
	}).ForEachT(func(call dtos.VariantCall) { // **
		accumulatedCalls = append(accumulatedCalls, &call)
	})

	// if len(accumulatedCalls) == 0 {
	// 	_t.Skip("No samples returned! Skipping --")
	// }

	for _, c := range accumulatedCalls {
		assert.NotEmpty(_t, c.SampleId)
		assert.NotEmpty(_t, c.GenotypeType)

		specificValidation(_t, c)
	}
}

func getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(_t *testing.T, includeInfo bool, sortByPosition c.SortDirection, genotype c.GenotypeQuery, referenceAllelePattern string, alternativeAllelePattern string) []dtos.VariantGetReponse {
	cfg := common.InitConfig()

	// retrieve the overview
	overviewJson := common.GetVariantsOverview(_t, cfg)

	// ensure the response is valid
	// TODO: error check instead of nil check
	assert.NotNil(_t, overviewJson)

	// generate all possible combinations of
	// available samples, assemblys, and chromosomes
	overviewCombinations := common.GetOverviewResultCombinations(overviewJson["chromosomes"], overviewJson["sampleIDs"], overviewJson["assemblyIDs"])

	// avoid overflow:
	// - shuffle all combinations and take top x
	x := 10
	croppedCombinations := make([][]string, len(overviewCombinations))
	perm := rand.Perm(len(overviewCombinations))
	for i, v := range perm {
		croppedCombinations[v] = overviewCombinations[i]
	}
	if len(croppedCombinations) > x {
		croppedCombinations = croppedCombinations[:x]
	}

	// initialize a common slice in which to
	// accumulate al responses asynchronously
	allDtoResponses := []dtos.VariantGetReponse{}
	allDtoResponsesMux := sync.RWMutex{}

	var combWg sync.WaitGroup
	for _, combination := range croppedCombinations {
		combWg.Add(1)
		go func(_wg *sync.WaitGroup, _combination []string) {
			defer _wg.Done()

			chrom := _combination[0]
			sampleId := _combination[1]
			assemblyId := a.CastToAssemblyId(_combination[2])

			// make the call
			dto := common.BuildQueryAndMakeGetVariantsCall(chrom, sampleId, includeInfo, sortByPosition, genotype, assemblyId, referenceAllelePattern, alternativeAllelePattern, "", false, _t, cfg)

			assert.Equal(_t, 1, len(dto.Results))

			// accumulate all response objects
			// to a common slice in an
			// asynchronous-safe manner
			allDtoResponsesMux.Lock()
			allDtoResponses = append(allDtoResponses, dto)
			allDtoResponsesMux.Unlock()
		}(&combWg, combination)
	}

	combWg.Wait()

	return allDtoResponses
}
