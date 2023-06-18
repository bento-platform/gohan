package common

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"gohan/api/models"
	c "gohan/api/models/constants"
	gq "gohan/api/models/constants/genotype-query"
	"gohan/api/models/dtos"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

const (
	VariantsOverviewPath                      string = "%s/variants/overview"
	VariantsGetBySampleIdsPathWithQueryString string = "%s/variants/get/by/sampleId%s"
	IngestionRequestsPath                     string = "%s/variants/ingestion/requests"
)

func InitConfig() *models.Config {
	var cfg models.Config

	// get this file's path
	_, filename, _, _ := runtime.Caller(0)
	folderpath := path.Dir(filename)

	// retrieve common's test.config
	f, err := os.Open(fmt.Sprintf("%s/test.config.yml", folderpath))
	if err != nil {
		processError(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		processError(err)
	}

	if cfg.Debug {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &cfg
}

func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}

func GetVariantsOverview(_t *testing.T, _cfg *models.Config) map[string]interface{} {
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

func BuildQueryAndMakeGetVariantsCall(
	chromosome string, sampleId string, includeInfo bool,
	sortByPosition c.SortDirection, genotype c.GenotypeQuery, assemblyId c.AssemblyId,
	referenceAllelePattern string, alternativeAllelePattern string, commaDeliminatedAlleles string,
	ignoreStatusCode bool, _t *testing.T, _cfg *models.Config) dtos.VariantGetReponse {

	queryString := fmt.Sprintf("?ids=%s&includeInfoInResultSet=%t&sortByPosition=%s&assemblyId=%s", sampleId, includeInfo, sortByPosition, assemblyId)

	if chromosome != "" {
		queryString = fmt.Sprintf("%s%s", queryString, fmt.Sprintf("&chromosome=%s", chromosome))
	}

	if genotype != gq.UNCALLED {
		queryString = fmt.Sprintf("%s%s", queryString, fmt.Sprintf("&genotype=%s", string(genotype)))
	}

	if referenceAllelePattern != "" {
		queryString = fmt.Sprintf("%s%s", queryString, fmt.Sprintf("&reference=%s", referenceAllelePattern))
	}
	if alternativeAllelePattern != "" {
		queryString = fmt.Sprintf("%s%s", queryString, fmt.Sprintf("&alternative=%s", alternativeAllelePattern))
	}
	if commaDeliminatedAlleles != "" {
		queryString = fmt.Sprintf("%s%s", queryString, fmt.Sprintf("&alleles=%s", commaDeliminatedAlleles))
	}
	url := fmt.Sprintf(VariantsGetBySampleIdsPathWithQueryString, _cfg.Api.Url, queryString)

	return makeGetVariantsCall(url, ignoreStatusCode, _t)
}

func makeGetVariantsCall(url string, ignoreStatusCode bool, _t *testing.T) dtos.VariantGetReponse {
	fmt.Printf("Calling %s\n", url)
	request, _ := http.NewRequest("GET", url, nil)

	client := &http.Client{}
	response, responseErr := client.Do(request)
	assert.Nil(_t, responseErr)

	defer response.Body.Close()

	if !ignoreStatusCode {
		// this test (at the time of writing) will only work if authorization is disabled
		shouldBe := 200
		assert.Equal(_t, shouldBe, response.StatusCode, fmt.Sprintf("Error -- Api GET %s Status: %s ; Should be %d", url, response.Status, shouldBe))
	}

	//	-- interpret array of ingestion requests from response
	respBody, respBodyErr := ioutil.ReadAll(response.Body)
	assert.Nil(_t, respBodyErr)

	//	--- transform body bytes to string
	respBodyString := string(respBody)

	//	-- convert to json and check for error
	var respDto dtos.VariantGetReponse
	jsonUnmarshallingError := json.Unmarshal([]byte(respBodyString), &respDto)
	assert.Nil(_t, jsonUnmarshallingError)

	return respDto
}
