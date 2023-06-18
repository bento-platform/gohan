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
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

const (
	VariantsOverviewPath                      string = "%s/variants/overview"
	VariantsGetBySampleIdsPathWithQueryString string = "%s/variants/get/by/sampleId%s"
	IngestionRequestsPath                     string = "%s/variants/ingestion/requests"

	DemoVcf1 string = `##fileformat=VCFv4.2
#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	S-1178-HAP
1	13656	.	CAG	C,<NON_REF>	868.60	.	BaseQRankSum=-5.505;DP=81;ExcessHet=3.0103;MLEAC=1,0;MLEAF=0.500,0.00;MQRankSum=-2.985;RAW_MQandDP=43993,81;ReadPosRankSum=-0.136	GT:AD:DP:GQ:PL:SB	0:50,25,0:75:99:876,0,2024,1026,2099,3126:4,46,5,20
10	28872481	.	CAAAA	C,CA,CAAA,CAAAAA,CAAAAAA,<NON_REF>	652.60	.	BaseQRankSum=0.029;DP=83;ExcessHet=3.0103;MLEAC=0,0,0,1,0,0;MLEAF=0.00,0.00,0.00,0.500,0.00,0.00;MQRankSum=-0.186;RAW_MQandDP=291409,83;ReadPosRankSum=-0.582	GT:AD:DP:GQ:PL:SB	0:19,3,2,5,29,9,0:67:99:660,739,2827,748,2714,2732,724,1672,1682,1587,0,340,338,249,265,321,956,929,699,245,898,866,1996,1991,1652,466,1006,1944:0,19,0,48
19	3619025	.	C	<NON_REF>	.	.	END=3619025	GT:DP:GQ:MIN_DP:PL	0:19:21:19:0,21,660
19	3619026	.	T	<NON_REF>	.	.	END=3619026	GT:DP:GQ:MIN_DP:PL	0:19:51:19:0,51,765`
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

func GetRootGohanPath() string {
	// check if file exists
	wd, err1 := os.Getwd()
	if err1 != nil {
		log.Println(err1)
	}
	fmt.Println(wd) // for example /home/user

	path := filepath.Dir(wd)
	for i := 1; i < 5; i++ {
		path = filepath.Dir(path)
	}

	return path
}

func IsError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}

func CreateAndGetNewFile(filePath string) (*os.File, error) {
	// - create file if not exists
	var (
		newFile    *os.File
		newFileErr error
	)

	_, newFileErr = os.Create(filePath)
	if IsError(newFileErr) {
		return nil, newFileErr
	}

	// - reopen file using READ & WRITE permission.
	newFile, newFileErr = os.OpenFile(filePath, os.O_RDWR, 0644)
	if IsError(newFileErr) {
		return nil, newFileErr
	}
	return newFile, newFileErr
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
