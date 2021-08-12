package api

import (
	"api/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	common "tests/common"

	"github.com/stretchr/testify/assert"
)

const (
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

	// -- insure it's an empty array
	assert.Equal(t, 0, len(ingestionRequestsRespJsonSlice))
}

func TestCanGetVariantsWithoutSamplesInResultset(t *testing.T) {
	cfg := common.InitConfig()

	dto := getVariantsWithSamplesInResultset(false, t, cfg)

	assert.Equal(t, 1, len(dto.Data))

	firstDataPointResults := dto.Data[0].Results
	assert.Nil(t, firstDataPointResults[0].Samples)
}

func TestCanGetVariantsWithSamplesInResultset(t *testing.T) {
	cfg := common.InitConfig()

	dto := getVariantsWithSamplesInResultset(true, t, cfg)

	assert.Equal(t, 1, len(dto.Data))

	firstDataPointResults := dto.Data[0].Results
	assert.NotNil(t, firstDataPointResults[0].Samples)
}

// Common utility functions for api tests
func getVariantsWithSamplesInResultset(includeSamples bool, _t *testing.T, _cfg *models.Config) models.VariantsResponseDTO {

	queryString := fmt.Sprintf("?chromosome=21&ids=hg00096&includeSamplesInResultSet=%t", includeSamples)
	url := fmt.Sprintf(VariantsGetBySampleIdsPathWithQueryString, _cfg.Api.Url, queryString)
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
