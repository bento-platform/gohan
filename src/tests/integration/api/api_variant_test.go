package api

import (
	"api/models"
	c "api/models/constants"
	a "api/models/constants/assembly-id"
	gq "api/models/constants/genotype-query"
	s "api/models/constants/sort"
	z "api/models/constants/zygosity"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	common "tests/common"
	testConsts "tests/common/constants"
	ratt "tests/common/constants/referenceAlternativeTestType"

	. "github.com/ahmetb/go-linq"

	"github.com/stretchr/testify/assert"
)

const (
	VariantsOverviewPath                      string = "%s/variants/overview"
	VariantsGetBySampleIdsPathWithQueryString string = "%s/variants/get/by/sampleId%s"
	IngestionRequestsPath                     string = "%s/variants/ingestion/requests"
)

func TestWithInvalidAuthenticationToken(t *testing.T) {
	cfg := common.InitConfig()

	request, _ := http.NewRequest("GET", cfg.Api.Url, nil)

	request.Header.Add("X-AUTHN-TOKEN", "gibberish")

	client := &http.Client{}
	response, responseErr := client.Do(request)
	assert.Nil(t, responseErr)

	defer response.Body.Close()

	// default response without a valid authentication token is is 401; consider it a pass
	var shouldBe int
	if cfg.AuthX.IsAuthorizationEnabled {
		shouldBe = 401
	} else {
		shouldBe = 200
	}

	assert.Equal(t, shouldBe, response.StatusCode, fmt.Sprintf("Error -- Api GET / Status: %s ; Should be %d", response.Status, shouldBe))
}

func TestVariantsOverview(t *testing.T) {
	cfg := common.InitConfig()

	overviewJson := getVariantsOverview(t, cfg)
	assert.NotNil(t, overviewJson)
}

func TestGetIngestionRequests(t *testing.T) {
	cfg := common.InitConfig()

	request, _ := http.NewRequest("GET", fmt.Sprintf(IngestionRequestsPath, cfg.Api.Url), nil)

	client := &http.Client{}
	response, responseErr := client.Do(request)
	assert.Nil(t, responseErr)

	defer response.Body.Close()

	// this test (at the time of writing) will only work if authorization is disabled
	shouldBe := 200
	assert.Equal(t, shouldBe, response.StatusCode, fmt.Sprintf("Error -- Api GET / Status: %s ; Should be %d", response.Status, shouldBe))

	//	-- interpret array of ingestion requests from response
	ingestionRequestsRespBody, ingestionRequestsRespBodyErr := ioutil.ReadAll(response.Body)
	assert.Nil(t, ingestionRequestsRespBodyErr)

	//	--- transform body bytes to string
	ingestionRequestsRespBodyString := string(ingestionRequestsRespBody)

	//	-- check for json error
	var ingestionRequestsRespJsonSlice []map[string]interface{}
	ingestionRequestsStringJsonUnmarshallingError := json.Unmarshal([]byte(ingestionRequestsRespBodyString), &ingestionRequestsRespJsonSlice)
	assert.Nil(t, ingestionRequestsStringJsonUnmarshallingError)

	// -- ensure the response is not nil
	assert.NotNil(t, len(ingestionRequestsRespJsonSlice))
}

func TestCanGetVariantsWithoutSamplesInResultset(t *testing.T) {

	allDtoResponses := getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(t, false, s.Undefined, gq.UNCALLED, "", "")

	// assert that all responses from all combinations have no results
	for _, dtoResponse := range allDtoResponses {
		firstDataPointResults := dtoResponse.Data[0].Results
		assert.Nil(t, firstDataPointResults[0].Samples)
	}
}

func TestCanGetVariantsWithSamplesInResultset(t *testing.T) {

	allDtoResponses := getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(t, true, s.Undefined, gq.UNCALLED, "", "")

	// assert that all of the responses include valid sample sets
	// - * accumulate all samples into a single list using the set of
	//   SelectManyT's and the SelectT
	// - ** iterate over each sample in the ForEachT
	var accumulatedSamples []*models.Sample

	From(allDtoResponses).SelectManyT(func(resp models.VariantsResponseDTO) Query { // *
		return From(resp.Data)
	}).SelectManyT(func(data models.VariantResponseDataModel) Query {
		return From(data.Results)
	}).SelectManyT(func(variant models.Variant) Query {
		return From(variant.Samples)
	}).SelectT(func(sample models.Sample) models.Sample {
		return sample
	}).ForEachT(func(sample models.Sample) { // **
		accumulatedSamples = append(accumulatedSamples, &sample)
	})

	if len(accumulatedSamples) == 0 {
		t.Skip("No samples returned! Skipping --")
	}

	for _, s := range accumulatedSamples {
		assert.NotEmpty(t, s.Id)
		assert.NotEmpty(t, s.Variation)
	}
}

func TestCanGetVariantsInAscendingPositionOrder(t *testing.T) {
	// retrieve responses in ascending order
	allDtoResponses := getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(t, false, s.Ascending, gq.UNCALLED, "", "")

	// assert the dto response slice is plentiful
	assert.NotNil(t, allDtoResponses)

	From(allDtoResponses).ForEachT(func(dto models.VariantsResponseDTO) {
		// ensure there is data
		assert.NotNil(t, dto.Data)

		// check the data
		From(dto.Data).ForEachT(func(d models.VariantResponseDataModel) {
			// ensure the variants slice is plentiful
			assert.NotNil(t, d.Results)

			latestSmallest := 0
			From(d.Results).ForEachT(func(dd models.Variant) {
				// verify order
				if latestSmallest != 0 {
					assert.True(t, latestSmallest <= dd.Pos)
				}

				latestSmallest = dd.Pos
			})
		})
	})
}

func TestCanGetVariantsInDescendingPositionOrder(t *testing.T) {
	// retrieve responses in descending order
	allDtoResponses := getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(t, false, s.Descending, gq.UNCALLED, "", "")

	// assert the dto response slice is plentiful
	assert.NotNil(t, allDtoResponses)

	From(allDtoResponses).ForEachT(func(dto models.VariantsResponseDTO) {
		// ensure there is data
		assert.NotNil(t, dto.Data)

		// check the data
		From(dto.Data).ForEachT(func(d models.VariantResponseDataModel) {
			// ensure the variants slice is plentiful
			assert.NotNil(t, d.Results)

			latestGreatest := 0
			From(d.Results).ForEachT(func(dd models.Variant) {
				if latestGreatest != 0 {
					assert.True(t, latestGreatest >= dd.Pos)
				}

				latestGreatest = dd.Pos
			})
		})
	})
}

func TestCanGetHeterozygousSamples(t *testing.T) {
	// trigger
	runAndValidateGenotypeQueryResults(t, gq.HETEROZYGOUS, validateHeterozygousSample)
}

func TestCanGetHomozygousReferenceSamples(t *testing.T) {
	// trigger
	runAndValidateGenotypeQueryResults(t, gq.HOMOZYGOUS_REFERENCE, validateHomozygousReferenceSample)
}

func TestCanGetHomozygousAlternateSamples(t *testing.T) {
	// trigger
	runAndValidateGenotypeQueryResults(t, gq.HOMOZYGOUS_ALTERNATE, validateHomozygousAlternateSample)
}

func TestCanGetHomozygousAlternateVariantsWithVariousReferences(t *testing.T) {
	// setup
	specificValidation := func(__t *testing.T, variant *models.Variant, referenceAllelePattern string, alternativeAllelePattern string) {
		// ensure test is formatted correctly
		assert.True(__t, alternativeAllelePattern == "")

		// validate variant
		assert.Contains(__t, variant.Ref, referenceAllelePattern)

		for _, sample := range variant.Samples {
			validateHomozygousAlternateSample(__t, &sample)
		}
	}

	executeReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HOMOZYGOUS_ALTERNATE, ratt.Reference, specificValidation)
}

func TestCanGetHomozygousReferenceVariantsWithVariousReferences(t *testing.T) {
	// setup
	specificValidation := func(__t *testing.T, variant *models.Variant, referenceAllelePattern string, alternativeAllelePattern string) {
		// ensure test is formatted correctly
		assert.True(__t, alternativeAllelePattern == "")

		// validate variant
		assert.Contains(__t, variant.Ref, referenceAllelePattern)

		for _, sample := range variant.Samples {
			validateHomozygousReferenceSample(__t, &sample)
		}
	}

	executeReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HOMOZYGOUS_REFERENCE, ratt.Reference, specificValidation)
}

func TestCanGetHeterozygousVariantsWithVariousReferences(t *testing.T) {
	// setup
	specificValidation := func(__t *testing.T, variant *models.Variant, referenceAllelePattern string, alternativeAllelePattern string) {
		// ensure test is formatted correctly
		assert.True(__t, alternativeAllelePattern == "")

		// validate variant
		assert.Contains(__t, variant.Ref, referenceAllelePattern)

		for _, sample := range variant.Samples {
			validateHeterozygousSample(__t, &sample)
		}
	}

	// trigger
	executeReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HETEROZYGOUS, ratt.Reference, specificValidation)
}

func TestCanGetHomozygousAlternateVariantsWithVariousAlternatives(t *testing.T) {
	// setup
	specificValidation := func(__t *testing.T, variant *models.Variant, referenceAllelePattern string, alternativeAllelePattern string) {
		// ensure test is formatted correctly
		assert.True(__t, referenceAllelePattern == "")

		// validate variant
		assert.Contains(__t, variant.Alt, alternativeAllelePattern)

		for _, sample := range variant.Samples {
			validateHomozygousAlternateSample(__t, &sample)
		}
	}

	// trigger
	executeReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HOMOZYGOUS_ALTERNATE, ratt.Alternative, specificValidation)
}

func TestCanGetHomozygousReferenceVariantsWithVariousAlternatives(t *testing.T) {
	// setup
	specificValidation := func(__t *testing.T, variant *models.Variant, referenceAllelePattern string, alternativeAllelePattern string) {
		// ensure test is formatted correctly
		assert.True(__t, referenceAllelePattern == "")

		// validate variant
		assert.Contains(__t, variant.Alt, alternativeAllelePattern)

		for _, sample := range variant.Samples {
			validateHomozygousReferenceSample(__t, &sample)
		}
	}

	// trigger
	executeReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HOMOZYGOUS_REFERENCE, ratt.Alternative, specificValidation)
}

func TestCanGetHeterozygousVariantsWithVariousAlternatives(t *testing.T) {
	// setup
	specificValidation := func(__t *testing.T, variant *models.Variant, referenceAllelePattern string, alternativeAllelePattern string) {
		// ensure test is formatted correctly
		assert.True(__t, referenceAllelePattern == "")

		// validate variant
		assert.Contains(__t, variant.Alt, alternativeAllelePattern)

		for _, sample := range variant.Samples {
			validateHeterozygousSample(__t, &sample)
		}
	}

	// trigger
	executeReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HETEROZYGOUS, ratt.Alternative, specificValidation)
}

// -- Common utility functions for api tests
func executeReferenceOrAlternativeQueryTestsOfVariousPatterns(_t *testing.T,
	genotypeQuery c.GenotypeQuery, refAltTestType testConsts.ReferenceAlternativeTestType,
	specificValidation func(__t *testing.T, variant *models.Variant, referenceAllelePattern string, alternativeAllelePattern string)) {

	// TODO: use some kind of Allele Enum
	patterns := []string{"A", "C", "T", "G"}
	var patWg sync.WaitGroup
	for _, pat := range patterns {
		patWg.Add(1)
		go func(_pat string, _patWg *sync.WaitGroup) {
			defer _patWg.Done()

			switch refAltTestType {
			case ratt.Reference:
				runAndValidateReferenceOrAlternativeQueryResults(_t, genotypeQuery, _pat, "", specificValidation)
			case ratt.Alternative:
				runAndValidateReferenceOrAlternativeQueryResults(_t, genotypeQuery, "", _pat, specificValidation)
			default:
				println("Skipping Test -- no Ref/Alt Test Type provided")
			}

		}(pat, &patWg)
	}
	patWg.Wait()
}

func runAndValidateReferenceOrAlternativeQueryResults(_t *testing.T,
	genotypeQuery c.GenotypeQuery,
	referenceAllelePattern string, alternativeAllelePattern string,
	specificValidation func(__t *testing.T, variant *models.Variant, referenceAllelePattern string, alternativeAllelePattern string)) {

	allDtoResponses := getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(_t, true, s.Undefined, genotypeQuery, referenceAllelePattern, alternativeAllelePattern)

	// assert that all of the responses include sample sets with the appropriate zygosity
	// - * accumulate all variants into a single list using the set of SelectManyT's and the SelectT
	// - ** iterate over each variant in the ForEachT
	var accumulatedVariants []*models.Variant

	From(allDtoResponses).SelectManyT(func(resp models.VariantsResponseDTO) Query { // *
		return From(resp.Data)
	}).SelectManyT(func(data models.VariantResponseDataModel) Query {
		return From(data.Results)
	}).SelectT(func(variant models.Variant) models.Variant {
		return variant
	}).ForEachT(func(variant models.Variant) { // **
		accumulatedVariants = append(accumulatedVariants, &variant)
	})

	if len(accumulatedVariants) == 0 {
		_t.Skip(fmt.Sprintf("No variants returned for patterns ref: '%s', alt: '%s'! Skipping --", referenceAllelePattern, alternativeAllelePattern))
	}

	for _, v := range accumulatedVariants {
		assert.NotNil(_t, v.Id)
		specificValidation(_t, v, referenceAllelePattern, alternativeAllelePattern)
	}

}

func runAndValidateGenotypeQueryResults(_t *testing.T, genotypeQuery c.GenotypeQuery, specificValidation func(__t *testing.T, sample *models.Sample)) {

	allDtoResponses := getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(_t, true, s.Undefined, genotypeQuery, "", "")

	// assert that all of the responses include heterozygous sample sets
	// - * accumulate all samples into a single list using the set of SelectManyT's and the SelectT
	// - ** iterate over each sample in the ForEachT
	var accumulatedSamples []*models.Sample

	From(allDtoResponses).SelectManyT(func(resp models.VariantsResponseDTO) Query { // *
		return From(resp.Data)
	}).SelectManyT(func(data models.VariantResponseDataModel) Query {
		return From(data.Results)
	}).SelectManyT(func(variant models.Variant) Query {
		return From(variant.Samples)
	}).SelectT(func(sample models.Sample) models.Sample {
		return sample
	}).ForEachT(func(sample models.Sample) { // **
		accumulatedSamples = append(accumulatedSamples, &sample)
	})

	if len(accumulatedSamples) == 0 {
		_t.Skip("No samples returned! Skipping --")
	}

	for _, s := range accumulatedSamples {
		assert.NotEmpty(_t, s.Id)
		assert.NotEmpty(_t, s.Variation)
		assert.NotEmpty(_t, s.Variation.Genotype)
		assert.NotEmpty(_t, s.Variation.Genotype.Zygosity)

		specificValidation(_t, s)
	}
}

func buildQueryAndMakeGetVariantsCall(chromosome string, sampleId string, includeSamples bool, sortByPosition c.SortDirection, genotype c.GenotypeQuery, assemblyId c.AssemblyId, referenceAllelePattern string, alternativeAllelePattern string, _t *testing.T, _cfg *models.Config) models.VariantsResponseDTO {

	queryString := fmt.Sprintf("?chromosome=%s&ids=%s&includeSamplesInResultSet=%t&sortByPosition=%s&assemblyId=%s", chromosome, sampleId, includeSamples, sortByPosition, assemblyId)

	if genotype != gq.UNCALLED {
		queryString = fmt.Sprintf("%s%s", queryString, fmt.Sprintf("&genotype=%s", string(genotype)))
	}

	if referenceAllelePattern != "" {
		queryString = fmt.Sprintf("%s%s", queryString, fmt.Sprintf("&reference=%s", referenceAllelePattern))
	}
	if alternativeAllelePattern != "" {
		queryString = fmt.Sprintf("%s%s", queryString, fmt.Sprintf("&alternative=%s", alternativeAllelePattern))
	}

	url := fmt.Sprintf(VariantsGetBySampleIdsPathWithQueryString, _cfg.Api.Url, queryString)

	return makeGetVariantsCall(url, _t)
}

func getVariantsOverview(_t *testing.T, _cfg *models.Config) map[string]interface{} {
	request, _ := http.NewRequest("GET", fmt.Sprintf(VariantsOverviewPath, _cfg.Api.Url), nil)

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
	chromosomesKey, ckOk := overviewRespJson["chromosomes"]
	assert.True(_t, ckOk)
	assert.NotNil(_t, chromosomesKey)

	variantIDsKey, vidkOk := overviewRespJson["variantIDs"]
	assert.True(_t, vidkOk)
	assert.NotNil(_t, variantIDsKey)

	sampleIDsKey, sidkOk := overviewRespJson["sampleIDs"]
	assert.True(_t, sidkOk)
	assert.NotNil(_t, sampleIDsKey)

	return overviewRespJson
}

func getOverviewResultCombinations(chromosomeStruct interface{}, sampleIdsStruct interface{}, assemblyIdsStruct interface{}) [][]string {
	var allCombinations = [][]string{}

	for i, _ := range chromosomeStruct.(map[string]interface{}) {
		for j, _ := range sampleIdsStruct.(map[string]interface{}) {
			for k, _ := range assemblyIdsStruct.(map[string]interface{}) {
				allCombinations = append(allCombinations, []string{i, j, k})
			}
		}
	}

	return allCombinations
}

func getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(_t *testing.T, includeSamples bool, sortByPosition c.SortDirection, genotype c.GenotypeQuery, referenceAllelePattern string, alternativeAllelePattern string) []models.VariantsResponseDTO {
	cfg := common.InitConfig()

	// retrieve the overview
	overviewJson := getVariantsOverview(_t, cfg)

	// ensure the response is valid
	// TODO: error check instead of nil check
	assert.NotNil(_t, overviewJson)

	// generate all possible combinations of
	// available samples, assemblys, and chromosomes
	overviewCombinations := getOverviewResultCombinations(overviewJson["chromosomes"], overviewJson["sampleIDs"], overviewJson["assemblyIDs"])

	// initialize a common slice in which to
	// accumulate al responses asynchronously
	allDtoResponses := []models.VariantsResponseDTO{}
	allDtoResponsesMux := sync.RWMutex{}

	var combWg sync.WaitGroup
	for _, combination := range overviewCombinations {
		combWg.Add(1)
		go func(_wg *sync.WaitGroup, _combination []string) {
			defer _wg.Done()

			chrom := _combination[0]
			sampleId := _combination[1]
			assemblyId := a.CastToAssemblyId(_combination[2])

			// make the call
			dto := buildQueryAndMakeGetVariantsCall(chrom, sampleId, includeSamples, sortByPosition, genotype, assemblyId, referenceAllelePattern, alternativeAllelePattern, _t, cfg)

			assert.Equal(_t, 1, len(dto.Data))

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

func makeGetVariantsCall(url string, _t *testing.T) models.VariantsResponseDTO {
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
	var respDto models.VariantsResponseDTO
	jsonUnmarshallingError := json.Unmarshal([]byte(respBodyString), &respDto)
	assert.Nil(_t, jsonUnmarshallingError)

	return respDto
}

// --- sample validation
func validateHeterozygousSample(__t *testing.T, sample *models.Sample) {
	assert.True(__t, sample.Variation.Genotype.Zygosity == z.Heterozygous)
}

func validateHomozygousReferenceSample(__t *testing.T, sample *models.Sample) {
	assert.True(__t, sample.Variation.Genotype.Zygosity == z.HomozygousReference)
}

func validateHomozygousAlternateSample(__t *testing.T, sample *models.Sample) {
	assert.True(__t, sample.Variation.Genotype.Zygosity == z.HomozygousAlternate)
}

// --
