package api

import (
	"api/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	common "tests/common"

	. "github.com/ahmetb/go-linq"

	"github.com/stretchr/testify/assert"
)

const (
	VariantsOverviewPath                      string = "%s/variants/overview"
	VariantsGetBySampleIdsPathWithQueryString string = "%s/variants/get/by/sampleId%s"
	IngestionRequestsPath                     string = "%s/variants/ingestion/requests"
)

func TestApiWithInvalidAuthenticationToken(t *testing.T) {
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

func TestApiVariantsOverview(t *testing.T) {
	cfg := common.InitConfig()

	overviewJson := getVariantsOverview(t, cfg)
	assert.NotNil(t, overviewJson)
}

func TestApiGetIngestionRequests(t *testing.T) {
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

	allDtoResponses := getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(t, false, "")

	// assert that all responses from all combinations have no results
	for _, dtoResponse := range allDtoResponses {
		firstDataPointResults := dtoResponse.Data[0].Results
		assert.Nil(t, firstDataPointResults[0].Samples)
	}
}

func TestCanGetVariantsWithSamplesInResultset(t *testing.T) {

	allDtoResponses := getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(t, true, "")

	// assert that all of the responses include valid sample sets
	// - * accumulate all samples into a single list using the set of
	//   SelectManyT's and the SelectT
	// - ** iterate over each sample in the ForEachT

	From(allDtoResponses).SelectManyT(func(resp models.VariantsResponseDTO) Query { // *
		return From(resp.Data)
	}).SelectManyT(func(data models.VariantResponseDataModel) Query {
		return From(data.Results)
	}).SelectManyT(func(variant models.Variant) Query {
		return From(variant.Samples)
	}).SelectT(func(sample models.Sample) models.Sample {
		return sample
	}).ForEachT(func(sample models.Sample) { // **
		assert.NotEmpty(t, sample.SampleId)
		assert.NotEmpty(t, sample.Variation)
	})
}

func TestCanGetVariantsInAscendingPositionOrder(t *testing.T) {
	// retrieve responses in ascending order
	allDtoResponses := getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(t, false, "asc")

	// assert the dto response slice is plentiful
	assert.NotNil(t, allDtoResponses)

	From(allDtoResponses).ForEachT(func(dto models.VariantsResponseDTO) {
		// assert there is data
		assert.NotNil(t, dto.Data)

		// check the data
		From(dto.Data).ForEachT(func(d models.VariantResponseDataModel) {
			// assert the variants slice is plentiful
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
	allDtoResponses := getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(t, false, "desc")

	// assert the dto response slice is plentiful
	assert.NotNil(t, allDtoResponses)

	From(allDtoResponses).ForEach(func(dtoInt interface{}) {
		dto := dtoInt.(models.VariantsResponseDTO)

		// assert there is data
		assert.NotNil(t, dto.Data)

		// check the data
		From(dto.Data).ForEach(func(dInt interface{}) {
			d := dInt.(models.VariantResponseDataModel)

			// assert the variants slice is plentiful
			assert.NotNil(t, d.Results)

			latestGreatest := 0

			From(d.Results).ForEach(func(ddInt interface{}) {
				dd := ddInt.(models.Variant)

				if latestGreatest != 0 {
					assert.True(t, latestGreatest >= dd.Pos)
				}

				latestGreatest = dd.Pos
			})
		})
	})
}

// -- Common utility functions for api tests
func buildQueryAndMakeGetVariantsCall(chromosome string, sampleId string, includeSamples bool, sortByPosition string, _t *testing.T, _cfg *models.Config) models.VariantsResponseDTO {

	if sortByPosition != "asc" && sortByPosition != "desc" {
		sortByPosition = "" // default to empty (will trigger ascending)
	}

	queryString := fmt.Sprintf("?chromosome=%s&ids=%s&includeSamplesInResultSet=%t&sortByPosition=%s", chromosome, sampleId, includeSamples, sortByPosition)
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

func getChromsAndSampleIDs(chromosomeStruct interface{}, sampleIdsStruct interface{}) [][]string {
	var allCombinations = [][]string{}

	for i, _ := range chromosomeStruct.(map[string]interface{}) {
		for j, _ := range sampleIdsStruct.(map[string]interface{}) {
			allCombinations = append(allCombinations, []string{i, j})
		}
	}

	return allCombinations
}

func getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(_t *testing.T, includeSamples bool, sortByPosition string) []models.VariantsResponseDTO {
	cfg := common.InitConfig()

	// todo: deduplicate
	overviewJson := getVariantsOverview(_t, cfg)
	assert.NotNil(_t, overviewJson)

	chromSampleIdCombinations := getChromsAndSampleIDs(overviewJson["chromosomes"], overviewJson["sampleIDs"])

	allDtoResponses := []models.VariantsResponseDTO{}
	for _, combination := range chromSampleIdCombinations {
		chrom := combination[0]
		sampleId := combination[1]

		dto := buildQueryAndMakeGetVariantsCall(chrom, sampleId, includeSamples, sortByPosition, _t, cfg)
		assert.Equal(_t, 1, len(dto.Data))
		allDtoResponses = append(allDtoResponses, dto)
	}

	return allDtoResponses
}

func makeGetVariantsCall(url string, _t *testing.T) models.VariantsResponseDTO {
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

// --
